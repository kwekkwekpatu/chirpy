FROM debian:stable-slim

# Create a public directory
RUN mkdir -p /public

# COPY source destination
COPY chirpy /bin/chirpy
COPY index.html /public/index.html

# Expose the port the app runs on
EXPOSE 8080
ENV PORT=8080

CMD ["/bin/chirpy"]
