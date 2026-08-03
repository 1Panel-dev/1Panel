export namespace Login {
    export interface ReqLoginForm {
        name: string;
        password: string;
        captcha: string;
        captchaID: string;
        authMethod: string;
        authSource: 'local' | 'ldap';
    }
    export interface MFALoginForm {
        sessionId: string;
        code: string;
        authMethod: string;
    }
    export interface ResLogin {
        name: string;
        role: string;
        token: string;
        mfaStatus: string;
        mfaSession: string;
    }
    export interface PasskeyBeginResponse {
        sessionId: string;
        publicKey: Record<string, any>;
    }
    export interface OIDCStatus {
        enabled: boolean;
        displayName: string;
        authorizationCode: boolean;
    }
    export interface LDAPStatus {
        enabled: boolean;
    }
    export interface OIDCBeginResponse {
        authorizationURL: string;
    }
    export interface OIDCFinishRequest {
        ticket: string;
    }
    export interface SAML2Status {
        enabled: boolean;
        displayName: string;
        syncLogout: boolean;
    }
    export interface SAML2RedirectNavigation {
        binding: 'redirect';
        redirectURL: string;
    }
    export interface SAML2PostNavigation {
        binding: 'post';
        postURL: string;
        fields: Record<string, string>;
    }
    export type SAML2Navigation = SAML2RedirectNavigation | SAML2PostNavigation;
    export interface SAML2BeginResponse {
        navigation: SAML2Navigation;
    }
    export interface SAML2FinishRequest {
        ticket: string;
    }
    export interface LogOutResponse {
        saml2Navigation?: SAML2Navigation;
    }
    export interface ResCaptcha {
        imagePath: string;
        captchaID: string;
        captchaLength: number;
    }
    export interface ResAuthButtons {
        [propName: string]: any;
    }

    export interface LoginSetting {
        isDemo: boolean;
        isIntl: boolean;
        isFxplay: boolean;
        language: string;
        menuTabs: string;
        menuAccordion: string;
        panelName: string;
        theme: string;
        isOffline: boolean;
        isEnterprise: boolean;
        needCaptcha: boolean;
        passkeySetting: boolean;
    }

    export interface AuthInfo {
        id: number;
        name: string;
        mfaStatus: string;
        mfaInterval: number;
        role: string;
        permissions: string[];
        masterOnlyPermissions?: string[];
        nodeRoles: Array<{ nodeId: number; nodeName: string; roleId: number; roleName: string }>;
        authSource: string;
        authSourceStatus: string;

        apiInterfaceStatus: string;
        apiKey: string;
        ipWhiteList: string;
        apiTrustedProxies: string;
        apiKeyValidityTime: number;
    }
    export interface AuthInfoUpdate {
        id: number;
        name: string;
        password: string;
        oldPassword: string;
    }
    export interface MFARequest {
        title: string;
        interval: number;
    }
    export interface MFAInfo {
        secret: string;
        qrImage: string;
    }
    export interface MFABind {
        secret: string;
        code: string;
        interval: number;
    }
    export interface ApiConfig {
        apiInterfaceStatus: string;
        apiKey: string;
        ipWhiteList: string;
        apiTrustedProxies: string;
        apiKeyValidityTime: number;
    }
    export interface PasswordUpdate {
        oldPassword: string;
        newPassword: string;
    }
}
