#!/usr/bin/env bash

#export MM_DEBUG=1
export MM_SERVICESETTINGS_SITEURL=http://127.0.0.1:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password

make deploy
