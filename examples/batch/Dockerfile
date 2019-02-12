FROM amazonlinux:latest

RUN yum -y install unzip aws-cli
ADD batch-job.sh /usr/local/bin/batch-job.sh
WORKDIR /tmp
USER nobody

ENTRYPOINT ["/usr/local/bin/batch-job.sh"]