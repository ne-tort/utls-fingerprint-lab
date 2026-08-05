# syntax=docker/dockerfile:1.7
FROM eclipse-temurin:21-jdk-jammy AS build
WORKDIR /app
COPY Main.java .
RUN javac Main.java

FROM eclipse-temurin:21-jre-jammy
WORKDIR /app
COPY --from=build /app/Main.class .
ENV TARGET_ID=java-jdk
ENV DIAL_HOST=capture:8443
ENTRYPOINT ["java", "Main"]
