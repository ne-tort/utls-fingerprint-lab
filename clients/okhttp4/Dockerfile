# syntax=docker/dockerfile:1.7
FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /src
COPY pom.xml .
RUN --mount=type=cache,target=/root/.m2 \
    mvn -q -DskipTests dependency:go-offline || true
COPY Main.java src/main/java/okhttp4/Main.java
RUN --mount=type=cache,target=/root/.m2 \
    mvn -q -DskipTests package

FROM eclipse-temurin:21-jre
COPY --from=build /src/target/okhttp4-1.0-jar-with-dependencies.jar /app.jar
ENTRYPOINT ["java", "-jar", "/app.jar"]
