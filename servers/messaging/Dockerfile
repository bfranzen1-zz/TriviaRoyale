FROM python:3-alpine
WORKDIR /app
COPY . /app
#RUN pip install --upgrade setuptools
#RUN pip3 install --upgrade pip
RUN pip3 install -r requirements.txt
EXPOSE 5000-5110
#CMD ["flask", "run", "--host", "0.0.0.0"]
ENTRYPOINT ["python3", "app.py"]