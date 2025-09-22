#!/bin/bash
# Docker remote environment setup for pm-instance.khost.dev

echo "🐳 Setting up Docker remote environment for pm-instance.khost.dev..."

export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH=$HOME/.docker/pm-instance
export DOCKER_HOST=tcp://pm-instance.khost.dev:2376

echo "✅ Docker environment configured!"
echo "   DOCKER_HOST: $DOCKER_HOST"
echo "   DOCKER_CERT_PATH: $DOCKER_CERT_PATH"
echo ""
echo "📝 You can now use Docker commands directly:"
echo "   docker ps"
echo "   docker run -d nginx"
echo "   docker-compose up"
echo ""
echo "🔧 To make this permanent, add to your shell profile:"
echo "   source $(pwd)/docker-remote-env.sh"