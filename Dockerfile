FROM phusion/baseimage
MAINTAINER Tamás Gulácsi <tgulacsi78@gmail.com>

# Disable SSH
RUN rm -rf /etc/service/sshd /etc/my_init.d/00_regen_ssh_host_keys.sh

RUN DEBIAN_FRONTEND=noninteractive apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y rst2pdf

# Clean up APT when done.
RUN DEBIAN_FRONTEND=noninteractive apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN useradd pdf
ENV HOME /home/pdf

## LOCAL
EXPOSE 2222:2222
CMD ["/sbin/my_init"]

RUN SD=/etc/service/rst2pdf-web; mkdir -p $SD && echo '#!/bin/sh -e\ncd /home/pdf\nexec ./bin/rst2pdf-web 2>&1' >$SD/run && chmod 0755 $SD/run
RUN LD=/var/log/rst2pdf-web; SD=/etc/service/rst2pdf-web/log; mkdir -p $LD $SD && echo 's10485760\nn10\nt86400\n!gzip -9c -' >$LD/config && echo "#!/bin/sh -e\nexec svlogd $DN" >$SD/run && chmod 0755 $SD/run

RUN mkdir -p $HOME/bin
ADD rst2pdf-web $HOME/bin/rst2pdf-web

# should run with -p 2222:2222
