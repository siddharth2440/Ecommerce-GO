FROM nginx:1.21.6

# default the defalt conf file
RUN rm /etc/nginx/conf.d/default.conf

# COpy our custom configuration file
COPY nginx.conf /etc/nginx/conf.d/default.conf

#this ensurs that  Nginx runs in the foreground,which is necessary for Docker
CMD [ "nginx","-g","daemon off;" ]
#Daemon-off = tells nginx to run in the foreground instead of background(daemon mode)