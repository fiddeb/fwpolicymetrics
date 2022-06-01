# fwpolicymetrics
A prometheus exporter for Juniper SRX dataplane logs shipped with kafka


> usage: go run main.go -bootstapservers "kafkahostname:9092" -group yourgroupname -topic yourtopicname


### metrics

The exporter produces metrics like these:

\# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
\# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 2.9649e-05
go_gc_duration_seconds{quantile="0.25"} 4.6984e-05
go_gc_duration_seconds{quantile="0.5"} 5.3559e-05
go_gc_duration_seconds{quantile="0.75"} 6.6915e-05
go_gc_duration_seconds{quantile="1"} 0.000317642
go_gc_duration_seconds_sum 0.093634037
go_gc_duration_seconds_count 1390
...
\# HELP kafka_messages_total How many messages recived from kafka, by patrtition
\# TYPE kafka_messages_total counter
kafka_messages_total{partition="0"} 183067
kafka_messages_total{partition="1"} 143800
kafka_messages_total{partition="2"} 269999
kafka_messages_total{partition="all"} 596866
\# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
\# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
\# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
\# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 5
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
\# HELP rt_flow_session_total How many policy decisions made by the firewall
\# TYPE rt_flow_session_total counter
rt_flow_session_total{action="close",destination="zone1",hostname="mydevice",policy="1",source="zone2"} 959
rt_flow_session_total{action="deny",destination="zone1",hostn3me="mydevice",policy="3",source="zone2"} 240
rt_flow_session_total{action="permit",destination="zone2",hostname="mydevice",policy="2",source="zone1"} 58

\# HELP syslog_messages_error_total How many messages failed in parser
\# TYPE syslog_messages_error_total counter
syslog_messages_error_total{type="rfc5424"} 76