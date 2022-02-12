FROM scratch

COPY ./bin/server /project_deflector/server

CMD ["/project_deflector/server"]
