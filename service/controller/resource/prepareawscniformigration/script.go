package prepareawscniformigration

import "fmt"

func getScript(cidr string) string {
	scr := `
set -o errexit
set -o nounset
set -o pipefail

CILIUM_CIDR=%s

lines="$(ip route show table all|grep "scope link" |grep -Po "dev \Keth[0-9*] table [0-9]+"|sed 's/ table /|/')"

while : ; do
  for line in $lines
  do
    ifname="$(echo "$line" | cut -d'|' -f1)"
    table="$(echo "$line" | cut -d'|' -f2)"

    (ip route show table $table | grep $CILIUM_CIDR) || (echo "Adding route for dev $ifname in table $table" && ip route add $CILIUM_CIDR dev cilium_host table $table)
  done

  sleep 5
done
`

	return fmt.Sprintf(scr, cidr)
}
