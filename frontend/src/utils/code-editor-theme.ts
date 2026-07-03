export type CodeEditorTheme = 'vs' | 'vs-dark' | 'hc-black';

export const codeEditorThemeStorageKey = 'code-theme';

export const getDefaultCodeEditorTheme = (isDarkTheme: boolean): CodeEditorTheme => {
    return isDarkTheme ? 'vs-dark' : 'vs';
};

export const resolveCodeEditorTheme = (savedTheme: string | null, isDarkTheme: boolean): CodeEditorTheme => {
    if (savedTheme === 'vs' || savedTheme === 'vs-dark' || savedTheme === 'hc-black') {
        return savedTheme;
    }
    return getDefaultCodeEditorTheme(isDarkTheme);
};
