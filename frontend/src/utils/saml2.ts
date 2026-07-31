export type SAML2Navigation =
    | {
          binding: 'redirect';
          redirectURL: string;
      }
    | {
          binding: 'post';
          postURL: string;
          fields: Record<string, string>;
      };

const validateNavigationURL = (value: string) => {
    // A freshly opened popup has an opaque `about:blank` origin after its
    // opener is detached. Resolve relative URLs against the 1Panel page
    // instead of that popup, otherwise `new URL()` receives "null" as base.
    const url = new URL(value, window.location.origin);
    if (!['http:', 'https:'].includes(url.protocol)) {
        throw new Error('Unsupported SAML2 navigation protocol');
    }
    return url.toString();
};

/**
 * Continue a backend-created SAML2 flow without interpreting or rewriting the
 * protocol payload in the browser.
 */
export const submitSAML2Navigation = (navigation: SAML2Navigation, targetWindow: Window = window) => {
    if (navigation.binding === 'redirect' && navigation.redirectURL) {
        targetWindow.location.assign(validateNavigationURL(navigation.redirectURL));
        return;
    }
    if (navigation.binding !== 'post' || !navigation.postURL || !navigation.fields) {
        throw new Error('Invalid SAML2 navigation response');
    }

    const form = targetWindow.document.createElement('form');
    form.method = 'POST';
    form.action = validateNavigationURL(navigation.postURL);
    form.style.display = 'none';

    Object.entries(navigation.fields || {}).forEach(([name, value]) => {
        const input = targetWindow.document.createElement('input');
        input.type = 'hidden';
        input.name = name;
        input.value = value;
        form.appendChild(input);
    });

    targetWindow.document.body.appendChild(form);
    form.submit();
    form.remove();
};
