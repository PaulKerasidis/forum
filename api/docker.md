# 1. Create network
docker network create forum-network

# 2. Navigate to API directory  
cd ~/Documents/forum

# 3. Build image
docker build -t api-image:latest .

# 4. Run container
docker run -d -p 8080:8080 -v api_db_data:/app/DBPath --name api-container --network forum-network api-image:latest



### check images and containers 

docker images 

docker ps 




## OTHER HELPER COMMANDS 
 # ==============================================
# DELETE CONTAINERS
# ==============================================

# Stop a running container
docker stop api-container

# Remove a stopped container
docker rm api-container

# Stop and remove in one command (force)
docker rm -f api-container

# Remove multiple containers
docker rm -f api-container frontend-container

# Remove ALL stopped containers
docker container prune

# ==============================================
# DELETE IMAGES
# ==============================================

# Remove a specific image
docker rmi api-image:latest

# Remove image by ID
docker rmi abc123def456

# Force remove image (even if containers are using it)
docker rmi -f api-image:latest

# Remove multiple images
docker rmi api-image:latest frontend-image:latest

# Remove all unused images
docker image prune

# Remove all images (WARNING: removes everything!)
docker rmi $(docker images -q)

# ==============================================
# COMBINED CLEANUP COMMANDS
# ==============================================

# Complete cleanup for API
docker rm -f api-container && docker rmi -f api-image:latest

# Complete cleanup for both API and Frontend
docker rm -f api-container frontend-container
docker rmi -f api-image:latest frontend-image:latest

# ==============================================
# CHECK WHAT EXISTS BEFORE DELETING
# ==============================================

# List all containers (running and stopped)
docker ps -a

# List all images
docker images

# List containers using specific image
docker ps -a --filter ancestor=api-image:latest

# ==============================================
# VOLUME AND NETWORK CLEANUP
# ==============================================

# Remove specific volume (WARNING: deletes database!)
docker volume rm api_db_data

# Remove all unused volumes
docker volume prune

# Remove specific network
docker network rm forum-network

# Remove all unused networks
docker network prune

# ==============================================
# NUCLEAR OPTION - REMOVE EVERYTHING
# ==============================================

# Remove ALL containers, images, volumes, networks
docker system prune -a --volumes

# Same but without confirmation prompt
docker system prune -a --volumes -f

# ==============================================
# SAFE CLEANUP SEQUENCE
# ==============================================

# 1. Stop containers
docker stop api-container frontend-container

# 2. Remove containers
docker rm api-container frontend-container

# 3. Remove images
docker rmi api-image:latest frontend-image:latest

# 4. Clean up unused resources
docker system prune

# ==============================================
# QUICK REFERENCE
# ==============================================

# Most common commands:
docker rm -f container-name     # Remove container (force)
docker rmi image-name:tag       # Remove image
docker ps -a                    # List all containers
docker images                   # List all im