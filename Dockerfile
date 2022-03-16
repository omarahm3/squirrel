FROM alpine

COPY squirreld /usr/bin/squirreld

RUN ls -la /usr/bin/squirreld

ENTRYPOINT [ "/usr/bin/squirreld" ]
