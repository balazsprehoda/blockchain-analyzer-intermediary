WHICHOS := $(shell uname)
GOPATH?=$(shell dirname $(shell dirname $(shell dirname ${PWD})))

basic:
	cd stack && make start
	cd network/basic && make start
	cd apps/dummyapp &&	sed -i -e "s/\"channelName\": \"fourchannel\"/\"channelName\": \"mychannel\"/g" config.json &&	sed -i -e "s/\"connection_profile\": \"..\/..\/network\/multichannel\/connectionProfile.json\"/\"connection_profile\": \"..\/..\/network\/basic\/connectionProfile.json\"/g" config.json
	make remove-intermediate
	cd apps/dummyapp && make install && make users && make invoke
	export PATH=${PATH}:${GOPATH}/bin && cd agent/fabricbeat && make update && make
	#Waiting for Kibana
	sleep 15
	cd agent/fabricbeat && NETWORK=basic ORG_NUMBER=1 PEER_NUMBER=0 ./fabricbeat -e -d "*"

multichannel:
	cd stack && make start
	cd network/multichannel && make start
	cd apps/dummyapp &&	sed -i -e "s/\"channelName\": \"mychannel\"/\"channelName\": \"fourchannel\"/g" config.json &&	sed -i -e "s/\"connection_profile\": \"..\/..\/network\/basic\/connectionProfile.json\"/\"connection_profile\": \"..\/..\/network\/multichannel\/connectionProfile.json\"/g" config.json
	make remove-intermediate
	cd apps/dummyapp && make install && make users && make invoke
	export PATH=${PATH}:${GOPATH}/bin && cd agent/fabricbeat && make update && make
	#Waiting for Kibana
	sleep 15
	cd agent/fabricbeat && NETWORK=multichannel ORG_NUMBER=1 PEER_NUMBER=0 ./fabricbeat -e -d "*" &
	cd agent/fabricbeat && NETWORK=multichannel ORG_NUMBER=1 PEER_NUMBER=1 ./fabricbeat -e -d "*" &
	cd agent/fabricbeat && NETWORK=multichannel ORG_NUMBER=2 PEER_NUMBER=0 ./fabricbeat -e -d "*"

destroy-basic:
	cd ./stack && make erase
	cd network/basic && make destroy

destroy-multichannel:
	cd ./stack && make erase
	cd network/multichannel && make destroy

remove-intermediate:
ifeq ($(WHICHOS),Darwin)
	cd apps/dummyapp && rm *-e
endif