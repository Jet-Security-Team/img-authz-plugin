#!/bin/sh

set -xe

sh -c "dockerd-entrypoint.sh &> /tmp/dockerd-logs &"

until docker ps
do
  echo "INFO: waiting for dockerd..."
  sleep 2
done

(cd ${PLUGIN_DIR} && PLUGIN_NAME="${1}" PLUGIN_TAG="${2}" make local_install)

kill `pidof dockerd`

sh -c "dockerd-entrypoint.sh --authorization-plugin=\"${1}:${2}\" &> /tmp/dockerd-logs &"

timeout 15 sh -c '''until docker ps
do
  echo "INFO: waiting for dockerd with plugin enabled..."
  sleep 2
done'''

set +ex

echo '''

██████╗ ██╗   ██╗███╗   ██╗███╗   ██╗██╗███╗   ██╗ ██████╗     ████████╗███████╗███████╗████████╗███████╗
██╔══██╗██║   ██║████╗  ██║████╗  ██║██║████╗  ██║██╔════╝     ╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██╔════╝
██████╔╝██║   ██║██╔██╗ ██║██╔██╗ ██║██║██╔██╗ ██║██║  ███╗       ██║   █████╗  ███████╗   ██║   ███████╗
██╔══██╗██║   ██║██║╚██╗██║██║╚██╗██║██║██║╚██╗██║██║   ██║       ██║   ██╔══╝  ╚════██║   ██║   ╚════██║
██║  ██║╚██████╔╝██║ ╚████║██║ ╚████║██║██║ ╚████║╚██████╔╝       ██║   ███████╗███████║   ██║   ███████║██╗██╗██╗
╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚═╝╚═╝  ╚═══╝ ╚═════╝        ╚═╝   ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚═╝╚═╝╚═╝

'''

cd ${PLUGIN_DIR}/test

export SHUNIT_COLOR="always"
shunit2 tests.sh

out=$?

if [ $out -ne 0 ]
then
  echo -e "\033[1m\e[1;31m  !! TESTS HAVE FAILED !!  Full Docker logs:\033[0m"
  cat /tmp/dockerd-logs
fi
