# database name
DB_NAME ?= postgres

# database type
DB_TYPE ?= postgres

# database username
DB_USER ?= postgres

# database password
DB_PWD ?= mysecretpassword

# psql URL
IP=127.0.0.1

PSQLURL ?= $(DB_TYPE)://$(DB_USER):$(DB_PWD)@$(IP):5432/$(DB_NAME)

.PHONY : postgresup postgresdown psql

postgresup:
	docker run --name test-postgres -v $(PWD):/usr/share/vw -e POSTGRES_PASSWORD=$(DB_PWD) -e PGTZ="+3" -p 5432:5432 -d $(DB_NAME)

postgresdown:
	docker stop test-postgres || true && docker rm test-postgres || true

psql:
	docker exec -it test-postgres psql $(PSQLURL)

