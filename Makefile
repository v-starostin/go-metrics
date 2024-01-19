.PHONY: run_db
run_db:
	@docker run \
    		-d \
    		-v `pwd`/db:/docker-entrypoint-initdb.d/ \
    		--rm \
    		-p 5432:5432 \
    		--name db \
    		-e POSTGRES_DB=metrics \
    		-e POSTGRES_USER=postgres \
    		-e POSTGRES_PASSWORD=postgres \
    		postgres:16