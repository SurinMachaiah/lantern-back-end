dev:
	docker-compose up

prod:
	docker-compose up -f ./docker-compose.yml

down:
	docker-compose down

down_clean:
	docker-compose down --rmi all -v

test:
	# run the go tests ignoring certain directory types.
	# -count=1 ensures no caching of test results
	# -covermode=count provides coverage information
	#
	# go test is provided the list of files to test using the `go list ./...` command.
	# the `go list ./...` command pipes its results into several grep functions that remove files that
	# don't need to be tested and will clog the results.
	#
	# grep commands
	# -E uses 'extended regular expressions'
	# -v says to ignore 
	#
	# in makefiles, '$' is a special character so need to be escaped, which is done using another '$'.
	go test -count=1 -covermode=count \
		`go list ./... \
		| grep -E -v "^github.com/onc-healthit/lantern-back-end/[a-zA-Z0-9]+/cmd" \
		| grep -E -v "^github.com/onc-healthit/lantern-back-end/lanternmq" \
		| grep -E -v "/mock$$"`