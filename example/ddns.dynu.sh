#!/bin/sh

# Copy this script to ddns.sh and edit the cresentials and hostname,
# So every time the development shell start it will register the dns
# Note: EC2 hosts don't change their IP while they are up, so registering when the container
# start guarantee it for the life of the container

set -e

# Modify the following two line for your specific test
credentials="<username>:<password>"
hostname="<full host name>"

curl "https://${credentials}@api.dynu.com/nic/update?hostname=${hostname}"
