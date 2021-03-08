# pip
参考https://github.com/influxdata/telegraf  
阉割版telegraf
自定义数据采集工具，利用golang channel的特性，把数据串起来，可以实现通过输入插件收集系统数据，统计数据，自定义数据等，流过Processor插件进行相关处理，最终通过输出插件，可将收集的数据发送到各种其他数据存储，服务和消息队列，包括InfluxDB，Graphite，OpenTSDB，Datadog，Librato，Kafka，MQTT，NSQ等等。
比如：在做统计的程序，通过输入插件，收集第三方系统的原始数据或者mysql中的数据，在processor插件中，可以处理图片落地等操作，在输出插件中，做统计数据落地等操作
比如：在考情统计系统中，通过输入插件，收集考勤人员的考勤规则，在输出插件做根据考勤规则生成考勤记录的操作
