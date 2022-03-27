FROM scratch

COPY ./bin/server /project_deflector/server
COPY ./env/.local.env /env/.local.env

CMD ["/project_deflector/server"]
