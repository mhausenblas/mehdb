FROM centos
WORKDIR /app
RUN chown -R 1001:1 /app
USER 1001
COPY mehdb .
EXPOSE 9876
CMD ["/app/mehdb"]
