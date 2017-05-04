package cloudconfig

type CompactTLSAssets struct {
	APIServerCA       string
	APIServerKey      string
	APIServerCrt      string
	WorkerCA          string
	WorkerKey         string
	WorkerCrt         string
	ServiceAccountCA  string
	ServiceAccountKey string
	ServiceAccountCrt string
	CalicoClientCA    string
	CalicoClientKey   string
	CalicoClientCrt   string
	EtcdServerCA      string
	EtcdServerKey     string
	EtcdServerCrt     string
}
