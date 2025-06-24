backend:
	cd api && go run main.go

frontend:
	cd frontend && go run main.go

dev:
	cd api && go run main.go & cd frontend && go run main.go