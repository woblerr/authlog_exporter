# Custom Maxmind database for tests

Create a custom Maxmind database for tests.

## Using docker

Simple
```bash
make docker-build-test-db
```
or
```bash
cd ./test_data
docker build -f Dockerfile.testdb -t authlog_exporter_build_test_db .
docker run -d --name=authlog_exporter_build_test_db authlog_exporter_build_test_db
docker cp authlog_exporter_build_test_db:/db/geolite2_test.mmdb ./
docker rm -f authlog_exporter_build_test_db
```
