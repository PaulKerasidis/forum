# 1. Navigate to frontend directory
cd /path/to/your/frontend

# 2. Build image
docker build -t frontend-image:latest .

# 3. Run container
docker run -d -p 3000:3000 --name frontend-container --network forum-network frontend-image:latest





# delete container
docker rm -f frontend-container

# delete image  
docker rmi frontend-image:latest

# complete cleanup 
docker rm -f frontend-container && docker rmi -f frontend-image:latest


## ================================ ##
##    CHECK THE NETWORK             ##
## ================================ ##

docker network inspect forum-network