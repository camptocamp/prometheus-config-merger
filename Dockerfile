FROM scratch
ADD prometheus-config-merger /
ENTRYPOINT ["/prometheus-config-merger"]
VOLUME [ "/etc/prometheus" ]
CMD [""]
