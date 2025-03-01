import { Request, Response, Router } from 'express';
import oClient, { BaseClient, errors, TokenSet } from 'openid-client';
import { EndSessionData } from '../../shared/interfaces/end-session-data';
import { SessionService } from './session';

const generators = oClient.generators; // else jest isn't working
const prefixPath = process.env.PREFIX_PATH;

/**
 * Build the root path. The exact path depends on the deployment & PREFIX_PATH value
 *
 * If PREFIX_PATH is defined, root location is set to <PREFIX_PATH>/bridge. Otherwise, root is set to / .
 *
 * Redirection to / will be either handled by Nginx (ex:- generic keptn deployment) OR the Express layer (ex:- local bridge development).
 */
function getRootLocation(): string {
  if (prefixPath !== undefined) {
    return `${prefixPath}/bridge`;
  }
  return '/';
}

function oauthRouter(
  client: BaseClient,
  redirectUri: string,
  logoutUri: string,
  reduceRefreshDateSeconds: number,
  session: SessionService
): Router {
  const router = Router();
  const additionalScopes = process.env.OAUTH_SCOPE ? ` ${process.env.OAUTH_SCOPE.trim()}` : '';
  const scope = `openid${additionalScopes}`;
  console.log('Using scope:', scope);

  /**
   * Router level middleware for login
   */
  router.get('/oauth/login', async (req: Request, res: Response) => {
    const codeVerifier = generators.codeVerifier();
    const codeChallenge = generators.codeChallenge(codeVerifier);
    const nonce = generators.nonce();
    const state = generators.state();
    try {
      await session.saveValidationData(state, codeVerifier, nonce);

      const authorizationUrl = client.authorizationUrl({
        scope,
        state,
        nonce,
        code_challenge: codeChallenge,
        code_challenge_method: 'S256',
      });
      res.redirect(authorizationUrl);
    } catch (e) {
      console.log(e);
    }
    return res;
  });

  /**
   * Router level middleware for redirect handling
   */
  router.get('/oauth/redirect', async (req: Request, res: Response) => {
    const params = client.callbackParams(req);
    const invalidRequest = {
      title: 'Internal error',
      message: 'Error while handling the redirect. Please retry and check whether the problem still exists.',
      location: getRootLocation(),
    };

    if (!params.code || !params.state) {
      return res.render('error', invalidRequest);
    }

    try {
      const validationData = await session.getAndRemoveValidationData(params.state);
      if (!validationData) {
        return res.render('error', invalidRequest);
      }

      const tokenSet = await client.callback(redirectUri, params, {
        code_verifier: validationData.codeVerifier,
        nonce: validationData.nonce,
        state: validationData._id,
        scope,
      });
      reduceRefreshDateBy(tokenSet, reduceRefreshDateSeconds);
      await session.authenticateSession(req, tokenSet);
      res.redirect(getRootLocation());
    } catch (error) {
      const err = error as errors.OPError | errors.RPError;
      console.log(`Error while handling the redirect. Cause : ${err.message}`);

      if (err.response?.statusCode === 403) {
        return res.render('error', {
          title: 'Permission denied',
          message:
            (err.response.body as Record<string, string>).message ?? 'User is not allowed to access the instance.',
        });
      } else {
        return res.render('error', invalidRequest);
      }
    }
  });

  /**
   * Router level middleware for logout
   */
  router.post('/oauth/logout', async (req: Request, res: Response) => {
    if (!session.isAuthenticated(req.session)) {
      // Session is not authenticated, redirect to root
      return res.json();
    }

    const hint = session.getLogoutHint(req) ?? '';
    if (req.session.tokenSet.access_token && client.issuer.metadata.revocation_endpoint) {
      client.revoke(req.session.tokenSet.access_token);
    }
    session.removeSession(req);

    if (client.issuer.metadata.end_session_endpoint) {
      const params: EndSessionData = {
        id_token_hint: hint,
        state: generators.state(),
        post_logout_redirect_uri: logoutUri,
        end_session_endpoint: client.issuer.metadata.end_session_endpoint,
      };
      return res.json(params);
    } else {
      return res.json();
    }
  });

  // exception for this endpoint (not "/oauth/" before) because this will be moved to the client in future
  router.get('/logoutsession', (req: Request, res: Response) => {
    return res.render('logout', { location: getRootLocation() });
  });

  return router;
}

/**
 * Sets the expiry date x seconds before the real one
 * @param tokenSet
 * @param seconds
 */
function reduceRefreshDateBy(tokenSet: TokenSet, seconds: number): void {
  tokenSet.expires_at = tokenSet.expires_at ? tokenSet.expires_at - seconds : undefined; // token should be refreshed x seconds earlier
}

export { oauthRouter, getRootLocation, reduceRefreshDateBy };
