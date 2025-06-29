# E2E Testing for superviz.io Install Command

Ce dossier contient les tests end-to-end pour la commande `install` de superviz.io.

## Architecture

Le système de tests e2e utilise Docker pour créer des conteneurs de différentes distributions Linux avec SSH activé, puis teste la commande `install` sur chacune d'entre elles.

### Distributions supportées

- **Ubuntu 22.04** - Port SSH 2201
- **Debian 12** - Port SSH 2202
- **Alpine 3.18** - Port SSH 2203
- **CentOS Stream 9** - Port SSH 2204
- **Fedora 39** - Port SSH 2205
- **Arch Linux** - Port SSH 2206

### Credentials par défaut

- **Username**: `testuser`
- **Password**: `testpass`
- **Root password**: `rootpass`

L'utilisateur `testuser` a les privilèges sudo sans mot de passe pour l'automatisation.

## Prérequis

1. **Docker**: Doit être installé et accessible
2. **Docker Compose**: Pour orchestrer les conteneurs
3. **Devcontainer**: Le devcontainer doit être configuré avec Docker-in-Docker

## Utilisation

### Via Makefile (recommandé)

```bash
# Setup de l'environnement e2e (build des images)
make e2e-setup

# Tests sur toutes les distributions
make e2e-test

# Test sur une distribution spécifique
make e2e-test-single DISTRO=ubuntu

# Nettoyage
make e2e-clean
```

### Scripts directs

```bash
# Test sur toutes les distributions
./test/e2e/run_e2e_tests.sh

# Test sur une distribution spécifique
./test/e2e/test_single_distro.sh ubuntu
```

## Configuration Docker

### Conteneurs individuels

Chaque distribution a son propre Dockerfile dans `test/e2e/docker/`:

- `ubuntu.Dockerfile` - Ubuntu 22.04 avec SSH
- `debian.Dockerfile` - Debian 12 avec SSH
- `alpine.Dockerfile` - Alpine 3.18 avec SSH
- `centos.Dockerfile` - CentOS Stream 9 avec SSH
- `fedora.Dockerfile` - Fedora 39 avec SSH
- `arch.Dockerfile` - Arch Linux avec SSH

### Orchestration

Le fichier `docker-compose.yml` démarre tous les conteneurs avec:

- SSH servers configurés
- Utilisateurs de test créés
- Ports mappés (2201-2206)
- Health checks activés

## Tests effectués

Pour chaque distribution, le test:

1. **Build** le binaire superviz
2. **Démarre** le conteneur Docker de la distribution
3. **Attend** que le service SSH soit prêt
4. **Execute** `superviz install --password=testpass --skip-host-key-check testuser@localhost`
5. **Vérifie** que:
   - La connexion SSH réussit
   - La distribution est détectée correctement
   - Le repository est configuré avec succès
   - Les commandes d'installation sont affichées

## Logs et Debug

Les logs de chaque test sont sauvegardés dans:

- `test_output_<distro>.log` - Sortie complète de la commande install

Pour deboguer un test qui échoue:

```bash
# Démarrer manuellement un conteneur
cd test/e2e/docker
docker-compose up -d ubuntu-test

# Se connecter manuellement pour debug
ssh -p 2201 -o StrictHostKeyChecking=no testuser@localhost
# (password: testpass)

# Tester manuellement
./test/e2e/test_single_distro.sh ubuntu
```

## Authentification

Le système supporte deux modes d'authentification:

### 1. Par mot de passe (utilisé dans les tests)

```bash
superviz install --password=testpass testuser@host
```

### 2. Par clé SSH

```bash
superviz install --ssh-key=~/.ssh/id_rsa testuser@host
```

## Limitations actuelles

1. **Docker requis**: Les tests nécessitent Docker-in-Docker dans le devcontainer
2. **Ports fixes**: Les ports SSH sont fixes (2201-2206)
3. **Credentials fixes**: Username/password sont codés en dur
4. **Réseau local**: Tests uniquement sur localhost

## Ajout d'une nouvelle distribution

Pour ajouter une nouvelle distribution:

1. **Créer** un nouveau Dockerfile dans `test/e2e/docker/`
2. **Ajouter** le service dans `docker-compose.yml`
3. **Mettre à jour** les configurations dans les scripts
4. **Ajouter** le handler de distribution si nécessaire

## Sécurité

⚠️ **Important**: Ces configurations sont uniquement pour les tests e2e:

- SSH avec mot de passe activé
- Host key verification désactivée
- Privilèges sudo sans mot de passe
- Credentials en dur

**Ne jamais utiliser en production !**
