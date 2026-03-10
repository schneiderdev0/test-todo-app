FROM rust:1.70-slim-buster AS builder
WORKDIR /app
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs && \
    cargo build --release --locked && \
    rm -rf src
COPY src ./src
RUN cargo build --release --locked

FROM debian:buster-slim
WORKDIR /app
RUN apt-get update && apt-get install -y \
    libssl1.1 \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/todo-rust ./
EXPOSE 8000
CMD ["./todo-rust"]