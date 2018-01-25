// ipam provides IP address management functionality.
package ipam

import (
	"bytes"
	"encoding/binary"
	"math"
	"net"
	"sort"

	"github.com/giantswarm/microerror"
)

// ipToDecimal converts a net.IP to an int.
func ipToDecimal(ip net.IP) int {
	t := ip
	if len(ip) == 16 {
		t = ip[12:16]
	}

	return int(binary.BigEndian.Uint32(t))
}

// decimalToIP converts an int to a net.IP.
func decimalToIP(ip int) net.IP {
	t := make(net.IP, 4)
	binary.BigEndian.PutUint32(t, uint32(ip))

	return t
}

// add increments the given IP by the number.
// e.g: add(10.0.4.0, 1) -> 10.0.4.1.
// Negative values are allowed for decrementing.
func add(ip net.IP, number int) net.IP {
	return decimalToIP(ipToDecimal(ip) + number)
}

// size takes a mask, and returns the number of addresses.
func size(mask net.IPMask) int {
	ones, _ := mask.Size()
	size := int(math.Pow(2, float64(32-ones)))

	return size
}

// newIPRange takes an IPNet, and returns the ipRange of the network.
func newIPRange(network net.IPNet) ipRange {
	start := network.IP
	end := add(network.IP, size(network.Mask)-1)

	return ipRange{start: start, end: end}
}

// freeIPRanges takes a network, and a list of subnets.
// It calculates available IPRanges, within the original network.
func freeIPRanges(network net.IPNet, subnets []net.IPNet) ([]ipRange, error) {
	freeSubnets := []ipRange{}
	networkRange := newIPRange(network)

	if len(subnets) == 0 {
		freeSubnets = append(freeSubnets, networkRange)
		return freeSubnets, nil
	}

	{
		// Check space between start of network and first subnet.
		firstSubnetRange := newIPRange(subnets[0])

		// Check the first subnet doesn't start at the start of the network.
		if !networkRange.start.Equal(firstSubnetRange.start) {
			// It doesn't, so we have a free range between the start
			// of the network, and the start of the first subnet.
			end := add(firstSubnetRange.start, -1)
			freeSubnets = append(freeSubnets,
				ipRange{start: networkRange.start, end: end},
			)
		}
	}

	{
		// Check space between each subnet.
		for i := 0; i < len(subnets)-1; i++ {
			currentSubnetRange := newIPRange(subnets[i])
			nextSubnetRange := newIPRange(subnets[i+1])

			// If the two subnets are not contiguous,
			if ipToDecimal(currentSubnetRange.end)+1 != ipToDecimal(nextSubnetRange.start) {
				// Then there is a free range between them.
				start := add(currentSubnetRange.end, 1)
				end := add(nextSubnetRange.start, -1)
				freeSubnets = append(freeSubnets, ipRange{start: start, end: end})
			}
		}
	}

	{
		// Check space between last subnet and end of network.
		lastSubnetRange := newIPRange(subnets[len(subnets)-1])

		// Check the last subnet doesn't end at the end of the network.
		if !lastSubnetRange.end.Equal(networkRange.end) {
			// It doesn't, so we have a free range between the end of the
			// last subnet, and the end of the network.
			start := add(lastSubnetRange.end, 1)
			freeSubnets = append(freeSubnets,
				ipRange{start: start, end: networkRange.end},
			)
		}
	}

	return freeSubnets, nil
}

// space takes a list of free ip ranges, and a mask,
// and returns the start IP of the first range that could fit the mask.
func space(freeIPRanges []ipRange, mask net.IPMask) (net.IP, error) {
	for _, freeIPRange := range freeIPRanges {
		start := ipToDecimal(freeIPRange.start)
		end := ipToDecimal(freeIPRange.end)

		if end-start+1 >= size(mask) {
			return freeIPRange.start, nil
		}
	}

	return nil, microerror.Maskf(spaceExhaustedError, "tried to fit: %v", mask)
}

// Free takes a network, a mask, and a list of subnets.
// An available network, within the first network, is returned.
func Free(network net.IPNet, mask net.IPMask, subnets []net.IPNet) (net.IPNet, error) {
	if size(network.Mask) < size(mask) {
		return net.IPNet{}, microerror.Maskf(
			maskTooBigError, "have: %v, requested: %v", network.Mask, mask,
		)
	}

	for _, subnet := range subnets {
		if !network.Contains(subnet.IP) {
			return net.IPNet{}, microerror.Maskf(
				ipNotContainedError, "%v is not contained by %v", subnet.IP, network,
			)
		}
	}

	sort.Sort(ipNets(subnets))

	// Find all the free IP ranges.
	freeIPRanges, err := freeIPRanges(network, subnets)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	// Attempt to find a free space, of the required size.
	freeIP, err := space(freeIPRanges, mask)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	// Invariant: The IP of the network returned should not be nil.
	if freeIP == nil {
		return net.IPNet{}, microerror.Mask(nilIPError)
	}

	freeNetwork := net.IPNet{IP: freeIP, Mask: mask}

	// Invariant: The IP of the network returned should be contained
	// within the network supplied.
	if !network.Contains(freeNetwork.IP) {
		return net.IPNet{}, microerror.Maskf(
			ipNotContainedError, "%v is not contained by %v", freeNetwork.IP, network,
		)
	}

	// Invariant: The mask of the network returned should be equal to
	// the mask supplied as an argument.
	if !bytes.Equal(mask, freeNetwork.Mask) {
		return net.IPNet{}, microerror.Maskf(
			maskIncorrectSizeError, "have: %v, requested: %v", freeNetwork.Mask, mask,
		)
	}

	return freeNetwork, nil
}

// Half takes a network and returns two subnets which split the network in
// half.
func Half(network net.IPNet) (first, second net.IPNet, err error) {
	ones, bits := network.Mask.Size()
	if ones == bits {
		return net.IPNet{}, net.IPNet{}, microerror.Maskf(maskTooBigError, "single IP mask %q is not allowed", network.Mask.String())
	}

	// Bit shift is dividing by 2.
	ones++
	mask := net.CIDRMask(ones, bits)

	// Compute first half.
	first, err = Free(network, mask, nil)
	if err != nil {
		return net.IPNet{}, net.IPNet{}, microerror.Mask(err)
	}

	// Second half is computed by getting next free.
	second, err = Free(network, mask, []net.IPNet{first})
	if err != nil {
		return net.IPNet{}, net.IPNet{}, microerror.Mask(err)
	}

	return first, second, nil
}
