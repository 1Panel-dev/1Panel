export function getAssetsFile(url: string) {
    return new URL(`../assets/apps/${url}`, import.meta.url).href;
}
