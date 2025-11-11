# Changelog - 11 novembre 2025

## Nouvelles fonctionnalités

### 1. Configuration de l'email du premier administrateur

**Variable d'environnement**: `FIRST_ADMIN_EMAIL`

Permet de définir automatiquement un utilisateur comme administrateur lors de sa première connexion.

**Configuration**:
```bash
FIRST_ADMIN_EMAIL=admin@example.com
```

**Implémentation**:
- Vérification au démarrage de l'application : si un utilisateur avec cet email existe déjà, il est marqué comme admin
- Vérification lors de la création d'un nouvel utilisateur via OIDC : si l'email correspond, l'utilisateur est créé en tant qu'admin
- Comparaison insensible à la casse

**Fichiers modifiés**:
- `backend/cmd/server/main.go` : Vérification au démarrage
- `backend/internal/api/auth_handler.go` : Vérification lors de la création d'utilisateur
- `.env.example` : Documentation de la variable

### 2. Configuration personnalisable des scopes OIDC

**Variable d'environnement**: `OIDC_SCOPES`

Permet de définir les scopes OIDC à demander lors de l'authentification.

**Format**: Liste de scopes séparés par des virgules

**Valeur par défaut**: `openid,email,groups`

**Exemples**:
```bash
# Configuration par défaut
OIDC_SCOPES=openid,email,groups

# Pour Keycloak sans scope groups
OIDC_SCOPES=openid,email,profile

# Avec scopes personnalisés
OIDC_SCOPES=openid,email,profile,roles
```

**Notes**:
- Le scope `openid` est toujours ajouté automatiquement s'il n'est pas présent
- Les scopes doivent être supportés par votre provider OIDC

**Implémentation**:
- Parsing de la variable d'environnement en tableau de scopes
- Validation et ajout automatique du scope `openid`
- Configuration dynamique de l'oauth2.Config

**Fichiers modifiés**:
- `backend/internal/auth/oidc.go` : Parsing et configuration des scopes
- `.env.example` : Documentation de la variable

### 3. Configuration personnalisable du claim des groupes OIDC

**Variable d'environnement**: `OIDC_GROUPS_CLAIM`

Permet de définir le nom du claim JWT qui contient les groupes de l'utilisateur.

**Valeur par défaut**: `groups`

**Exemples**:
```bash
# Configuration par défaut
OIDC_GROUPS_CLAIM=groups

# Pour Keycloak avec realm roles
OIDC_GROUPS_CLAIM=realm_access.roles

# Pour Keycloak avec client roles
OIDC_GROUPS_CLAIM=resource_access.librecov.roles
```

**Formats supportés**:
- Tableau de chaînes : `["admin", "developer"]`
- Chaîne unique : `"admin"`
- Claims imbriqués avec notation pointée : `realm_access.roles`

**Implémentation**:
- Extraction flexible des groupes depuis n'importe quel claim JWT
- Support des claims imbriqués via notation pointée
- Gestion de plusieurs formats (array, string)
- Stockage en JSON dans la base de données

**Fichiers modifiés**:
- `backend/internal/auth/oidc.go` : Extraction des groupes depuis claim configurable
- `.env.example` : Documentation de la variable

## Corrections de bugs

### 1. Liste des projets non visible après connexion

**Problème**: La liste des projets n'était pas visible après connexion, bien que l'utilisateur soit authentifié.

**Cause**: Les handlers de l'API ne vérifiaient pas correctement si l'utilisateur était authentifié (ignoraient la valeur `exists` retournée par le middleware).

**Solution**: Ajout de vérifications explicites de l'existence de l'utilisateur dans tous les handlers de projets.

**Fichiers modifiés**:
- `backend/internal/api/project_handler.go` : Ajout de `exists` check dans les 8 méthodes (List, Get, Create, Update, Delete, GetShares, CreateShare, DeleteShare)

### 2. Table project_shares manquante

**Problème**: Erreur SQL "relation project_shares does not exist" lors de l'accès aux projets.

**Cause**: Le modèle `ProjectShare` existait mais n'était pas inclus dans l'AutoMigrate.

**Solution**: Ajout de `&models.ProjectShare{}` dans la liste des modèles à migrer automatiquement.

**Fichiers modifiés**:
- `backend/internal/database/database.go` : Ajout de ProjectShare dans AutoMigrate

## Documentation

### 1. Configuration OIDC complète

**Fichier**: `OIDC_CONFIGURATION.md`

Documentation complète pour configurer OIDC avec différents providers:
- Keycloak (realm roles et group mapper)
- Azure AD
- Okta
- Auth0

Inclut:
- Exemples de configuration par provider
- Formats supportés pour le claim groups
- Dépannage des erreurs courantes
- Tests de configuration

### 2. Mise à jour du README

Ajout d'une section sur les nouvelles variables OIDC dans le README principal avec lien vers la documentation détaillée.

## Configuration recommandée pour Keycloak

Si vous utilisez Keycloak, voici la configuration recommandée :

```bash
# Dans votre .env
OIDC_ISSUER=https://keycloak.example.com/realms/master
OIDC_CLIENT_ID=librecov
OIDC_REDIRECT_URL=http://localhost:4000/auth/callback

# Scopes (sans 'groups' car non supporté par défaut par Keycloak)
OIDC_SCOPES=openid,email,profile

# Utiliser les realm roles comme groupes
OIDC_GROUPS_CLAIM=realm_access.roles

# Premier administrateur
FIRST_ADMIN_EMAIL=admin@example.com

# Frontend
FRONTEND_URL=http://localhost:4000/
```

### Configuration Keycloak requise

1. **Créer des realm roles** (ex: admin, developer, viewer)
2. **Assigner les roles aux utilisateurs**
3. **Dans le client OIDC**, activer "Add to ID token" pour les realm roles

## Tests effectués

✅ Authentification OIDC avec Keycloak  
✅ Extraction des groupes depuis `realm_access.roles`  
✅ Liste des projets visible après connexion  
✅ Partage de projets par groupe fonctionnel  
✅ Toutes les tables créées automatiquement  
✅ Premier administrateur configuré automatiquement  

## Notes de migration

Si vous mettez à jour depuis une version précédente :

1. **Ajouter les nouvelles variables d'environnement** dans votre `.env` :
   ```bash
   OIDC_SCOPES=openid,email,profile
   OIDC_GROUPS_CLAIM=realm_access.roles
   FIRST_ADMIN_EMAIL=admin@example.com
   ```

2. **Reconstruire l'image Docker** :
   ```bash
   docker compose down
   docker compose up -d --build
   ```

3. **Vérifier les logs** pour s'assurer qu'il n'y a pas d'erreurs :
   ```bash
   docker logs librecov-librecov-1 --tail 50
   ```

4. **Tester l'authentification** :
   - Se connecter via OIDC
   - Vérifier que les groupes sont correctement extraits
   - Vérifier que la liste des projets est visible

## Compatibilité

- Go 1.25+
- PostgreSQL 16
- Keycloak 22+ (testé avec master realm)
- Autres providers OIDC (Azure AD, Okta, Auth0) supportés mais non testés

## Prochaines étapes recommandées

1. **Tester avec d'autres providers OIDC** (Azure AD, Okta, Auth0)
2. **Ajouter des tests unitaires** pour les nouvelles fonctionnalités
3. **Documenter l'API de partage de projets** dans la documentation Swagger/OpenAPI
4. **Ajouter une interface utilisateur** pour gérer les partages de projets directement depuis le frontend
