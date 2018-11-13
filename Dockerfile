ARG PREFIX=golang
FROM ${PREFIX}:1.11.2-stretch
ARG ARCH=x86_64
ENV arch=${ARCH}
RUN apt-get update && apt-get install -y libgtk2.0-dev
USER ${username}
CMD setarch ${arch} ./release.sh
