# Enhanced Version Command

La commande `version` de superviz.io a été améliorée pour offrir plus de flexibilité avec différents formats de sortie.

## Fonctionnalités

### Formats de sortie multiples

La commande version supporte 4 formats de sortie différents :

#### Format par défaut (human-readable)

```bash
svz version
```

```text
Version:       v0.0.3-9-g790018b-dirty
Commit:        790018b
Built at:      2025-06-29T12:02:58Z
Built by:      vscode
Go version:    go1.24.3
OS/Arch:       linux/arm64
```

#### Format JSON

```bash
svz version --format=json
```

```json
{
  "version": "v0.0.3-9-g790018b-dirty",
  "commit": "790018b",
  "built_at": "2025-06-29T12:02:58Z",
  "built_by": "vscode",
  "go_version": "go1.24.3",
  "os_arch": "linux/arm64"
}
```

#### Format YAML

```bash
svz version --format=yaml
```

```yaml
version: v0.0.3-9-g790018b-dirty
commit: 790018b
built_at: 2025-06-29T12:02:58Z
built_by: vscode
go_version: go1.24.3
os_arch: linux/arm64
```

#### Format court

```bash
svz version --format=short
```

```text
v0.0.3-9-g790018b-dirty
```

## Flags disponibles

- `-f, --format string`: Format de sortie (default|json|yaml|short) (défaut: "default")
- `-h, --help`: Afficher l'aide

## Build avec informations de version

Pour injecter les bonnes informations de version lors du build, utilisez le script fourni :

```bash
./scripts/build.sh [nom_binaire]
```

Ce script injecte automatiquement :

- Version (depuis les tags Git)
- Hash du commit Git
- Date et heure de build
- Utilisateur qui a fait le build
- Version de Go utilisée
- Architecture cible

## Optimisations de performance

La commande version implémente plusieurs optimisations de performance suivant les règles de codage du projet :

### Zero-allocation patterns

- Structures pré-allouées pour JSON/YAML
- Écriture directe des bytes au lieu de fmt.Fprintf
- Réutilisation des encodeurs

### Memory optimization

- Ordre des champs optimisé pour la mémoire
- Singleton pattern pour l'instance par défaut
- Structures compactes

### I/O optimization

- Écriture directe des bytes au lieu de fmt.Fprintf
- Minimisation des syscalls

## Cas d'usage

### Scripts de CI/CD

```bash
# Obtenir juste la version pour les tags
VERSION=$(svz version --format=short)

# Obtenir toutes les infos en JSON pour logging
svz version --format=json | jq .
```

### Debugging

```bash
# Voir les informations complètes
svz version
```

### Intégration dans d'autres outils

```bash
# Export des données en YAML pour traitement
svz version --format=yaml > version.yaml
```

## Architecture technique

La commande version suit l'architecture modulaire du projet :

- **Command layer** (`internal/cli/commands/version`): Interface CLI avec Cobra
- **Service layer** (`internal/services`): Logique métier et formatage
- **Provider layer** (`internal/providers`): Accès aux données de version

Cette séparation permet :

- Tests unitaires isolés pour chaque couche
- Injection de dépendances pour les tests
- Réutilisation du service dans d'autres contextes
- Maintien du singleton pattern pour les performances
