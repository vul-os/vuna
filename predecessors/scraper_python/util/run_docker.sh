PERSISIT_PATH=/home/imran/Documents/personal/projects/mongo
docker stop mongo_db
docker rm mongo_db
docker run -it -d --name mongo_db -v $PERSISIT_PATH:/data/db -p 27017:27017 mongo:latest