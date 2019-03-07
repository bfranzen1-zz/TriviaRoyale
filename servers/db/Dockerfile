FROM mysql
COPY schema.sql /docker-entrypoint-initdb.d/schema.sql
ENTRYPOINT ["--default-authentication-plugin=mysql_native_password"]