package configs

type config struct {
	Etcd etcd
}

func NewConfig() *config {
	return &config{Etcd:
	  etcd{
	    Endpoints: []string{
	    	"", // todo 写etcd地址
	    },
	  },
	}
}
