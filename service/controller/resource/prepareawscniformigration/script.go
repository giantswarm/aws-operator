package prepareawscniformigration

import "fmt"

func getScript(cidr string) string {
	scr := `
set -o errexit
set -o nounset
set -o pipefail

DEVICE=cilium_host
CILIUM_CIDR=%s

while ! ip a show dev $DEVICE
do 
  echo "Waiting for device $DEVICE to exist."
  sleep 5
done

lines="$(ip route show table all|grep "scope link"|grep -E "dev eth[0-9]+ table [0-9]+"|awk '{ print $3 "|" $5 }')"

while : ; do
  for line in $lines
  do
    ifname="$(echo "$line" | cut -d'|' -f1)"
    table="$(echo "$line" | cut -d'|' -f2)"

    (ip route show table $table | grep $CILIUM_CIDR) || (echo "Adding route for dev $ifname in table $table" && ip route add $CILIUM_CIDR dev $DEVICE table $table)
  done

  sleep 5
done
`

	return fmt.Sprintf(scr, cidr)
}
