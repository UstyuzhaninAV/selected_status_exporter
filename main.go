package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SelectelStatus struct {
	Result struct {
		StatusOverall struct {
			Updated    time.Time `json:"updated"`
			Status     string    `json:"status"`
			StatusCode int       `json:"status_code"`
		} `json:"status_overall"`
		Status []struct {
			ID         string    `json:"id"`
			Name       string    `json:"name"`
			Updated    time.Time `json:"updated"`
			Status     string    `json:"status"`
			StatusCode int       `json:"status_code"`
			Containers []struct {
				ID         string    `json:"id"`
				Name       string    `json:"name"`
				Updated    time.Time `json:"updated"`
				Status     string    `json:"status"`
				StatusCode int       `json:"status_code"`
			} `json:"containers"`
		} `json:"status"`
		Incidents   []interface{} `json:"incidents"`
		Maintenance struct {
			Active   []interface{} `json:"active"`
			Upcoming []struct {
				Name                 string    `json:"name"`
				ID                   string    `json:"_id"`
				DatetimeOpen         time.Time `json:"datetime_open"`
				DatetimePlannedStart time.Time `json:"datetime_planned_start"`
				DatetimePlannedEnd   time.Time `json:"datetime_planned_end"`
				Messages             []struct {
					Details  string    `json:"details"`
					State    int       `json:"state"`
					Status   int       `json:"status"`
					Datetime time.Time `json:"datetime"`
				} `json:"messages"`
				ContainersAffected []struct {
					Name string `json:"name"`
					ID   string `json:"_id"`
				} `json:"containers_affected"`
				ComponentsAffected []struct {
					Name string `json:"name"`
					ID   string `json:"_id"`
				} `json:"components_affected"`
			} `json:"upcoming"`
		} `json:"maintenance"`
	} `json:"result"`
}


func main() {
	log.Println("Selectel Status Exporter запущен")

	http.Handle("/metrics", promhttp.Handler())

	go recordMetrics()

	srv := &http.Server{
		Addr: "0.0.0.0:80",

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,

		Handler: nil,
	}

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	log.Println("Экспортер готов принимать запросы от прометеуса на /metrics")

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	<-c

	log.Println("Изящно завершаем работу экспортера...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
func initGauges() map[string]prometheus.Gauge {
	selectelNames := make(map[string][]string)
	selectelStructFields := []string{"id", "name", "updated", "status", "status_code"}
	selectelNames["Status"] = selectelStructFields
	selectelNames["Containers"] = selectelStructFields
	selectelNames["StatusOverall"] = []string{"updated", "status", "status_code"}


	promGauges := make(map[string]prometheus.Gauge)

	for name, fields := range selectelNames {
		for _, field := range fields {
			promGauges["selectel_status_"+name+"_"+field] = prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "selectel_status_" + name + "_" + field,
				Help: "selectel status " + name + " " + field,
			})
		}
	}

	for _, v := range promGauges {
		prometheus.MustRegister(v)
	}
	return promGauges
}

func recordMetrics() {
	gauge := initGauges()
	for {
		s := SelectelStatusResponce{}

		// делаем запрос к Selectel API
		if err := getSelectelStatus(&s); err != nil {
			log.Printf("Не удалось получить ответ от Selectel API! Ошибка: %v\n", err)
			continue
		}

		// status_overall
		gauge["selectel_status_status_overall_updated"].Set(float64(s.Data.result.status_overall.updated))
		gauge["selectel_status_status_overall_status"].Set(float64(s.Data.result.status_overall.status))
		gauge["selectel_status_status_overall_status_code"].Set(float64(s.Data.result.status_overall.status_code))

		// status
		gauge["selectel_status_status_id"].Set(float64(s.Data.result.status.id))
	  gauge["selectel_status_status_name"].Set(float64(s.Data.result.status.name))
		gauge["selectel_status_status_updated"].Set(float64(s.Data.result.status.updated))
		gauge["selectel_status_status_status"].Set(float64(s.Data.result.status.status))
		gauge["selectel_status_status_status_code"].Set(float64(s.Data.result.status.status_code))
		// gauge["selectel_status_status_containers"].Set(float64(s.Data.result.status.containers))


		time.Sleep(time.Hour * 1)
	}
}

func getSelectelStatus(selectelMetrics *SelectelStatusResponce) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://selectel.status.io/1.0/status/5980813dd537a2a7050004bd", nil)
	if err != nil {
		return err
	}

	temp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(temp, &selectelMetrics); err != nil {
		return err
	}

	return nil
}
