up:
	docker compose up -d

down:
	docker compose down --remove-orphans --volumes

topic_create:
	docker compose exec -it kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create \
		--topic $(topic) \
		--bootstrap-server localhost:9092 \
		--partitions 3 \
		--replication-factor 1

topic_describe:
	docker compose exec -it kafka /opt/bitnami/kafka/bin/kafka-topics.sh --describe \
		--topic $(topic) \
		--bootstrap-server localhost:9092

psql:
	docker compose exec pg psql -U postgres

redis:
	docker compose exec redis redis-cli

start:
	lsof -ti :8080,8081,8082 | xargs kill -9 &
	cd feed-svc && go run main.go &
	cd follow-svc && go run main.go &
	cd post-svc && go run main.go &
	cd feed-generation-svc && go run main.go &
	wait