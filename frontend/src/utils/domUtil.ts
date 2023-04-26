// create style element
export const createStyleElement = (css: string) => {
    const style = document.createElement('style');
    const textNode = document.createTextNode(css);
    style.appendChild(textNode);
    return style;
};
