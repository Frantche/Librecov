# Configuration OIDC - Scopes et Groups

## Variables d'environnement OIDC

### OIDC_SCOPES

Définit les scopes OIDC à demander lors de l'authentification.

**Format**: Liste de scopes séparés par des virgules

**Valeur par défaut**: `openid,email,groups`

**Exemples**:
```bash
# Configuration par défaut (avec groups)
OIDC_SCOPES=openid,email,groups

# Configuration minimale
OIDC_SCOPES=openid,email

# Configuration avec scopes personnalisés
OIDC_SCOPES=openid,email,profile,roles

# Configuration Keycloak avec roles
OIDC_SCOPES=openid,email,profile
```

**Note**: Le scope `openid` est toujours ajouté automatiquement s'il n'est pas présent.

### OIDC_GROUPS_CLAIM

Définit le nom du claim JWT qui contient les groupes de l'utilisateur.

**Format**: Chaîne de caractères

**Valeur par défaut**: `groups`

**Exemples**:
```bash
# Configuration par défaut
OIDC_GROUPS_CLAIM=groups

# Pour Keycloak avec realm roles
OIDC_GROUPS_CLAIM=realm_access.roles

# Pour Keycloak avec client roles
OIDC_GROUPS_CLAIM=resource_access.librecov.roles

# Pour Azure AD
OIDC_GROUPS_CLAIM=groups

# Pour Auth0
OIDC_GROUPS_CLAIM=https://example.com/groups
```

## Configuration par provider

### Keycloak

#### Option 1: Utiliser les realm roles (recommandé)

1. **Configuration Keycloak**:
   - Créer des realm roles (ex: `admin`, `developer`, `viewer`)
   - Assigner les roles aux utilisateurs
   - Dans le client OIDC, activer "Add to ID token" pour les realm roles

2. **Configuration Librecov**:
```bash
OIDC_SCOPES=openid,email,profile
OIDC_GROUPS_CLAIM=realm_access.roles
```

#### Option 2: Utiliser un group mapper

1. **Configuration Keycloak**:
   - Créer des groupes
   - Dans le client OIDC, ajouter un mapper de type "Group Membership"
   - Configurer le Token Claim Name à `groups`
   - Activer "Add to ID token"

2. **Configuration Librecov**:
```bash
OIDC_SCOPES=openid,email,profile
OIDC_GROUPS_CLAIM=groups
```

### Azure AD

```bash
OIDC_SCOPES=openid,email,profile
OIDC_GROUPS_CLAIM=groups
```

**Note**: Dans Azure AD, vous devez également configurer les "Group claims" dans votre application registration.

### Okta

```bash
OIDC_SCOPES=openid,email,profile,groups
OIDC_GROUPS_CLAIM=groups
```

### Auth0

```bash
OIDC_SCOPES=openid,email,profile
OIDC_GROUPS_CLAIM=https://example.com/groups
```

**Note**: Auth0 requiert des namespaced claims. Configurez une Rule ou Action pour ajouter les groupes.

## Formats supportés pour le claim groups

Le claim groups peut être dans différents formats:

1. **Tableau de chaînes** (recommandé):
```json
{
  "groups": ["admin", "developer", "project-viewers"]
}
```

2. **Chaîne unique**:
```json
{
  "groups": "admin"
}
```

3. **Claim imbriqué** (Keycloak realm roles):
```json
{
  "realm_access": {
    "roles": ["admin", "developer"]
  }
}
```

**Note**: Pour les claims imbriqués, utilisez la notation pointée dans `OIDC_GROUPS_CLAIM`.

## Exemple de configuration complète

### .env ou docker-compose.yml

```bash
# Configuration OIDC
OIDC_ISSUER=https://keycloak.example.com/realms/master
OIDC_CLIENT_ID=librecov
OIDC_REDIRECT_URL=http://localhost:4000/auth/callback

# Configuration des scopes (ajustez selon votre provider)
OIDC_SCOPES=openid,email,profile
OIDC_GROUPS_CLAIM=realm_access.roles

# Configuration admin
FIRST_ADMIN_EMAIL=admin@example.com

# Cookies
COOKIE_DOMAIN=
COOKIE_SECURE=false

# Frontend
FRONTEND_URL=http://localhost:4000
```

## Dépannage

### Erreur: "Invalid scopes"

Si vous voyez cette erreur dans les logs:
```
error=invalid_scope&error_description=Invalid+scopes%3A+...
```

**Solution**: Vérifiez que votre provider OIDC supporte les scopes demandés.

1. Vérifiez la configuration de votre client OIDC dans votre provider
2. Ajustez `OIDC_SCOPES` pour n'inclure que les scopes supportés
3. Pour Keycloak, retirez le scope `groups` si vous n'avez pas configuré de mapper de groupes

Exemple pour Keycloak sans groups:
```bash
OIDC_SCOPES=openid,email,profile
OIDC_GROUPS_CLAIM=realm_access.roles
```

### Les groupes ne sont pas extraits

1. **Vérifiez le token JWT**:
   - Utilisez jwt.io pour décoder votre token
   - Vérifiez que le claim contenant les groupes est présent
   - Notez le nom exact du claim

2. **Ajustez OIDC_GROUPS_CLAIM**:
   - Utilisez le nom exact du claim trouvé dans le token
   - Pour les claims imbriqués, utilisez la notation pointée

3. **Vérifiez les logs**:
```bash
docker logs librecov-librecov-1 --tail 50
```

### Test de configuration

Pour tester si vos groupes sont correctement extraits, après connexion:

1. Appelez l'endpoint `/auth/me`:
```bash
curl http://localhost:4000/auth/me \
  -H "Cookie: session_id=YOUR_SESSION_ID"
```

2. Vérifiez que le champ `groups` contient vos groupes:
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "User Name",
  "groups": "[\"admin\",\"developer\"]",
  "admin": true
}
```

## Partage de projets avec les groupes

Une fois les groupes configurés correctement, vous pouvez:

1. **Créer un projet** (propriétaire uniquement)
2. **Partager avec un groupe**:
   - Aller dans les paramètres du projet
   - Section "Group Shares"
   - Entrer le nom du groupe
   - Le projet sera accessible à tous les utilisateurs ayant ce groupe

Les utilisateurs ayant le groupe pourront:
- Voir le projet dans leur liste
- Voir les détails du projet
- Voir les builds et la couverture

Seul le propriétaire peut:
- Modifier le projet
- Supprimer le projet
- Gérer les partages
