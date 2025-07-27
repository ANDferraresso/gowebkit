#!/bin/bash

# Interrompi se un comando fallisce
set -e

# Chiedi la versione
echo "Inserisci la versione da taggare (es: v1.0.0):"
read VERSION

# Passaggio su develop
echo "🔁 Passaggio a develop..."
git checkout develop

# Push degli ultimi cambiamenti su develop
echo "⬆️  Push delle modifiche su develop..."
git push origin develop

# Passaggio a main
echo "🔁 Passaggio a main..."
git checkout main

# Merge di develop in main
echo "🔀 Merge di develop in main..."
git merge --no-ff develop

# Build/test (opzionale, puoi commentare se non ti serve)
echo "🧪 Esecuzione build e test..."
go build
go test ./...

# Push su main
echo "⬆️  Push su main..."
git push origin main

# Creazione del tag
echo "🏷️  Creazione tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

# Push del tag
echo "⬆️  Push del tag $VERSION..."
git push origin "$VERSION"

# Ritorno su develop
git checkout develop

echo "✅ Release $VERSION completata!"
