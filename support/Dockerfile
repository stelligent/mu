FROM java:8

ENV no_proxy=169.254.169.254,172.17.0.1,127.0.0.1
ENV NO_PROXY=169.254.169.254,172.17.0.1,127.0.0.1

RUN apt-get update && apt-get install -y python-pip
RUN pip install -U boto boto3

RUN mkdir -p /app
ADD ecs-env.py /app/ecs-env.py
ADD my-app.jar /app/app.jar

ENV SERVER_PORT 8080 
EXPOSE 8080
ENTRYPOINT [ "/app/ecs-env.py", "java $JAVA_OPTS -Djava.security.egd=file:/dev/./urandom -jar /app/app.jar" ]
