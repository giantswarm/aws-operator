package certs

type File struct {
	AbsolutePath string
	Data         []byte
}

type Files []File

func NewFilesCluster(cluster Cluster) Files {
	common := newFilesClusterCommon(cluster)
	master := newFilesClusterMaster(cluster)
	worker := newFilesClusterWorker(cluster)

	all := Files{}
	all = append(all, common...)
	all = append(all, master...)
	all = append(all, worker...)

	return all
}

func NewFilesClusterMaster(cluster Cluster) Files {
	common := newFilesClusterCommon(cluster)
	master := newFilesClusterMaster(cluster)

	all := Files{}
	all = append(all, common...)
	all = append(all, master...)

	return all
}

func NewFilesClusterWorker(cluster Cluster) Files {
	common := newFilesClusterCommon(cluster)
	worker := newFilesClusterWorker(cluster)

	all := Files{}
	all = append(all, common...)
	all = append(all, worker...)

	return all
}

func newFilesClusterCommon(cluster Cluster) Files {
	return Files{
		// Calico client.
		{
			AbsolutePath: "/etc/kubernetes/ssl/calico/client-ca.pem",
			Data:         cluster.CalicoClient.CA,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/calico/client-crt.pem",
			Data:         cluster.CalicoClient.Crt,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/calico/client-key.pem",
			Data:         cluster.CalicoClient.Key,
		},
		// Etcd client.
		// TODO create separate etcd client certificates.
		{
			AbsolutePath: "/etc/kubernetes/ssl/etcd/client-ca.pem",
			Data:         cluster.EtcdServer.CA,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/etcd/client-crt.pem",
			Data:         cluster.EtcdServer.Crt,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/etcd/client-key.pem",
			Data:         cluster.EtcdServer.Key,
		},
	}
}

func newFilesClusterMaster(cluster Cluster) Files {
	return Files{
		// Kubernetes API server.
		{
			AbsolutePath: "/etc/kubernetes/ssl/apiserver-ca.pem",
			Data:         cluster.APIServer.CA,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/apiserver-crt.pem",
			Data:         cluster.APIServer.Crt,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/apiserver-key.pem",
			Data:         cluster.APIServer.Key,
		},
		// Etcd server.
		{
			AbsolutePath: "/etc/kubernetes/ssl/etcd/server-ca.pem",
			Data:         cluster.EtcdServer.CA,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/etcd/server-crt.pem",
			Data:         cluster.EtcdServer.Crt,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/etcd/server-key.pem",
			Data:         cluster.EtcdServer.Key,
		},
		// Service account.
		{
			AbsolutePath: "/etc/kubernetes/ssl/service-account-ca.pem",
			Data:         cluster.ServiceAccount.CA,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/service-account-crt.pem",
			Data:         cluster.ServiceAccount.Crt,
		},
		{
			AbsolutePath: "/etc/kubernetes/ssl/service-account-key.pem",
			Data:         cluster.ServiceAccount.Key,
		},
	}
}

func newFilesClusterWorker(cluster Cluster) Files {
	return Files{
		// Kubernetes worker.
		{
			Data:         cluster.Worker.CA,
			AbsolutePath: "/etc/kubernetes/ssl/worker-ca.pem",
		},
		{
			Data:         cluster.Worker.Crt,
			AbsolutePath: "/etc/kubernetes/ssl/worker-crt.pem",
		},
		{
			Data:         cluster.Worker.Key,
			AbsolutePath: "/etc/kubernetes/ssl/worker-key.pem",
		},
	}
}
