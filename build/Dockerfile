# eg: docker build --target xxx --platform linux/amd64 --build-arg VER=${VER} -f Dockerfile -t xxx:${VER} .
ARG VER

FROM busybox:stable-glibc as csm-prometheus-collector
LABEL version="${VER}"
LABEL maintainers="Huawei CSM development team"
LABEL description="Kubernetes CSM(prometheus) for Huawei Storage"

ARG binary=./csm-prometheus-collector
COPY ${binary} csm-prometheus-collector
ENTRYPOINT ["/csm-prometheus-collector"]


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM busybox:stable-glibc as csm-cmi
LABEL version="${VER}"
LABEL maintainers="Huawei CSM development team"
LABEL description="Kubernetes CSM(cmi) for Huawei Storage"

ARG binary=./csm-cmi
COPY ${binary} csm-cmi
ENTRYPOINT ["/csm-cmi"]


FROM busybox:stable-glibc as csm-topo-service
LABEL version="${VER}"
LABEL maintainers="Huawei CSM development team"
LABEL description="Kubernetes CSM(topo) for Huawei Storage"

ARG binary=./csm-topo-service
COPY ${binary} csm-topo-service
ENTRYPOINT ["/csm-topo-service"]

FROM busybox:stable-glibc as csm-liveness-probe
LABEL version="${VER}"
LABEL maintainers="Huawei CSM development team"
LABEL description="Kubernetes CSM(livenessprobe) for Huawei Storage"

ARG binary=./csm-liveness-probe
COPY ${binary} csm-liveness-probe
ENTRYPOINT ["/csm-liveness-probe"]