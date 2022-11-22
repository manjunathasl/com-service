# com-service
how to run 

clone repository navigate to root dirctory of the project and follow the steps mentioned below

1. using docker -> 

    docker build -t comm-app .
    
    docker run comm-app


2. using docker-compose ->

    docker-compose up
   
external packages used 

github.com/shirou/gopsutil/v3: used only for finding os process to send the SIGINT signal.
