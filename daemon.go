package rachek

import (
	"encoding/json"
	"fmt"
	"github.com/michaelklishin/rabbit-hole"
	"github.com/pdedkov/goconfig"
	"net/http"
)

type Config struct {
	Url      string
	User     string
	Password string
	Queues   []struct {
		Vhost string `toml:"name""`
		Queue []struct {
			Queue        string  `toml:"title"`
			ErrorLevel   float32 `toml:"error"`
			WarningLevel float32 `toml:"warning"`
			Consumers    float32
			Messages     float32
		} `toml:"queue"`
	} `toml:"vhosts"`
}

type Daemon struct {
	Client *rabbithole.Client
	Config *Config
}

func (d *Daemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var res struct {
		Status  bool     `json:"status"`
		Error   []string `json:"error"`
		Warning []string `json:"warning"`
	}

	for _, q := range d.Config.Queues {
		items, err := d.Client.ListQueuesIn(q.Vhost)
		if err != nil {
			fmt.Printf("%+v", err)
			continue
		}
		if len(items) == 0 {
			res.Error = append(res.Error, fmt.Sprintf("Empty Vhost %s", q.Vhost))
			continue
		}
	LOOP:
		for indx, rec := range q.Queue {
			for _, i := range items {
				if rec.Queue == i.Name {
					q.Queue[indx].Consumers = float32(i.Consumers)
					q.Queue[indx].Messages = float32(i.MessagesReady + i.MessagesUnacknowledged)

					if q.Queue[indx].Messages > q.Queue[indx].Consumers*q.Queue[indx].ErrorLevel {
						res.Error = append(res.Error, fmt.Sprintf("%s:%s %.f %.f", q.Vhost, i.Name, q.Queue[indx].Messages, q.Queue[indx].Consumers))
					} else if q.Queue[indx].Messages > q.Queue[indx].Consumers*q.Queue[indx].WarningLevel {
						res.Warning = append(res.Warning, fmt.Sprintf("%s:%s %.f %.f", q.Vhost, i.Name, q.Queue[indx].Messages, q.Queue[indx].Consumers))
					}
					continue LOOP
				}
			}
			res.Error = append(res.Error, fmt.Sprintf("queue %s not found in vhost %s", rec.Queue, q.Vhost))
		}
	}
	if len(res.Error) == 0 && len(res.Warning) == 0 {
		res.Status = true
	}
	json, _ := json.Marshal(res)
	w.Write(json)
}

// NewDaemon create new Daemon
func NewDaemon(config string) (*Daemon, error) {
	conf := &Config{}
	err := goconfig.NewConfigFromFile(config, conf)

	if err != nil {
		return nil, err
	}

	cl, err := rabbithole.NewClient(conf.Url, conf.User, conf.Password)
	if err != nil {
		return nil, err
	}
	return &Daemon{cl, conf}, nil
}
