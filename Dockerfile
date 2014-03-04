FROM phusion/baseimage
MAINTAINER Tamás Gulácsi <tgulacsi78@gmail.com>

# Disable SSH
RUN rm -rf /etc/service/sshd /etc/my_init.d/00_regen_ssh_host_keys.sh

RUN DEBIAN_FRONTEND=noninteractive apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y rst2pdf

# Clean up APT when done.
RUN DEBIAN_FRONTEND=noninteractive apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN SD=/etc/service/rst2pdf-web; mkdir -p $SD && echo '#!/bin/sh -e\ncd /home/pdf\nexec setuser pdf ./bin/rst2pdf-web -hostport=:22221 2>&1' >$SD/run && chmod 0755 $SD/run
RUN LD=/var/log/rst2pdf-web; SD=/etc/service/rst2pdf-web/log; mkdir -p $LD $SD && echo 's10485760\nn10\nt86400\n!gzip -9c -' >$LD/config && echo "#!/bin/sh -e\nexec svlogd $DN" >$SD/run && chmod 0755 $SD/run

RUN useradd pdf
ENV HOME /home/pdf

RUN mkdir -p $HOME/bin 
ADD rst2pdf-web $HOME/bin/rst2pdf-web
RUN chown pdf: $HOME

## LOCAL
EXPOSE 22221:22221
CMD ["/sbin/my_init"]

# should run with -p 22221:22221
