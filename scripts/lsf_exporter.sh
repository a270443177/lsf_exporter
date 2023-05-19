#!/bin/bash
#
#       /etc/rc.d/init.d/lsf_exporter
#
# chkconfig: 2345 80 80
#
# config: /etc/prometheus/lsf_exporter.conf
# pidfile: /var/run/lsf_exporter.pid
 
# Source function library.
. /etc/init.d/functions
 
 
RETVAL=0
PROG="lsf_exporter"
#DAEMON_SYSCONFIG=/etc/sysconfig/${PROG}
DAEMON=/usr/bin/${PROG} #要把安装目录下/opt/lsf_exporter/lsf_exporter可执行文件拷贝到/usr/bin目录下
PID_FILE=/var/run/${PROG}.pid
LOCK_FILE=/var/lock/subsys/${PROG}
LOG_FILE=/var/log/lsf_exporter.log
DAEMON_USER="root"
FQDN=$(hostname)
GOMAXPROCS=$(grep -c ^processor /proc/cpuinfo)
ARGS=""
PORT=9818
#. ${DAEMON_SYSCONFIG}
 

start() {
  if check_status > /dev/null; then
    echo "lsf_exporter is already running"
    exit 0
  fi
 
  echo -n $"Starting lsf_exporter: "
  daemonize -u ${DAEMON_USER} -p ${PID_FILE} -l ${LOCK_FILE} -a -e ${LOG_FILE} -o ${LOG_FILE} ${DAEMON} ${ARGS}
  RETVAL=$?
  echo ""
  return $RETVAL
}
 
stop() {
    echo -n $"Stopping node_exporter: "
    killproc -p ${PID_FILE} -d 10 ${DAEMON}
    RETVAL=$?
    echo
    [ $RETVAL = 0 ] && rm -f ${LOCK_FILE} ${PID_FILE}
    return $RETVAL
}
 
check_status() {
    status -p ${PID_FILE} ${DAEMON}
    RETVAL=$?
    return $RETVAL
}
 
case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    status)
        check_status
        ;;
    reload|force-reload)
        reload
        ;;
    restart)
        stop
        start
        ;;
    *)
        N=/etc/init.d/${NAME}
        echo "Usage: $N {start|stop|status|restart|force-reload}" >&2
        RETVAL=2
        ;;
esac
 
exit ${RETVAL}