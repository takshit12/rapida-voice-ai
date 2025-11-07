FROM tiangolo/uvicorn-gunicorn:latest

ENV HOME=/app
ENV APP_HOME=$HOME
WORKDIR $APP_HOME

# Create app user
RUN addgroup --system rapida-app && adduser --system --group rapida-app
RUN mkdir -p $APP_HOME && chown rapida-app:rapida-app $APP_HOME

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

# Install system build dependencies for pillow-heif
USER root
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    automake \
    autoconf \
    libtool \
    pkg-config \
    cmake \
    git \
    wget \
    curl \
    libheif-dev \
    libde265-dev \
    libjpeg-dev \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY --chown=rapida-app:rapida-app api/document-api/requirements.txt requirements.txt

# Install Python dependencies (no --user)
RUN pip install --upgrade pip
RUN pip install -r requirements.txt

# Switch to non-root user for app runtime
USER rapida-app

# Copy app code
COPY --chown=rapida-app:rapida-app api/document-api/app ./app

EXPOSE 9010

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9010"]
