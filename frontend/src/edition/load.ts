import communityProvider from './community';
import type { EditionFrontendProvider } from './provider';
import proProvider from '@xpack-pro/edition';
import eeProvider from '@xpack-ee/edition';

const edition = (import.meta.env.VITE_FRONTEND_EDITION || 'community').toLowerCase();

export function loadEditionProvider(): EditionFrontendProvider {
    switch (edition) {
        case 'pro':
            return proProvider;
        case 'ee':
            return eeProvider;
        default:
            return communityProvider;
    }
}
