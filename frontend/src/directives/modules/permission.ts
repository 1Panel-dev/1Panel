import type { Directive, DirectiveBinding, VNode } from 'vue';
import {
    hasManagePermissionAccess,
    hasPermissionAccess,
    type PermissionBindingValue,
    type PermissionMode,
} from '@/utils/permission';

type PermissionControlledComponent = {
    setPermissionDisabled?: (disabled: boolean) => void;
};

const PERMISSION_DISABLED_ATTR = 'data-permission-disabled';
const PERMISSION_POINTER_EVENTS_ATTR = 'data-permission-pointer-events';
const PERMISSION_NATIVE_DISABLED_ATTR = 'data-permission-native-disabled';
const PERMISSION_TABINDEX_ATTR = 'data-permission-tabindex';
const NODE_ADMIN_PERMISSION_ATTR = 'data-node-admin-permission';

const getDisableTargets = (el: HTMLElement) => {
    const targets = [el, ...Array.from(el.querySelectorAll<HTMLElement>('button, input, select, textarea'))];
    return Array.from(new Set(targets));
};

const disableNativeControls = (el: HTMLElement) => {
    for (const target of getDisableTargets(el)) {
        if (
            target instanceof HTMLButtonElement ||
            target instanceof HTMLInputElement ||
            target instanceof HTMLSelectElement ||
            target instanceof HTMLTextAreaElement
        ) {
            if (!target.hasAttribute(PERMISSION_NATIVE_DISABLED_ATTR)) {
                target.setAttribute(PERMISSION_NATIVE_DISABLED_ATTR, target.disabled ? 'true' : 'false');
            }
            target.disabled = true;
            continue;
        }

        if (!target.hasAttribute(PERMISSION_TABINDEX_ATTR)) {
            const tabindex = target.getAttribute('tabindex');
            target.setAttribute(PERMISSION_TABINDEX_ATTR, tabindex ?? '');
        }
        target.setAttribute('tabindex', '-1');
    }
};

const enableNativeControls = (el: HTMLElement) => {
    for (const target of getDisableTargets(el)) {
        if (
            target instanceof HTMLButtonElement ||
            target instanceof HTMLInputElement ||
            target instanceof HTMLSelectElement ||
            target instanceof HTMLTextAreaElement
        ) {
            const previousDisabled = target.getAttribute(PERMISSION_NATIVE_DISABLED_ATTR);
            if (previousDisabled !== null) {
                target.disabled = previousDisabled === 'true';
                target.removeAttribute(PERMISSION_NATIVE_DISABLED_ATTR);
            }
            continue;
        }

        const previousTabindex = target.getAttribute(PERMISSION_TABINDEX_ATTR);
        if (previousTabindex !== null) {
            if (previousTabindex === '') {
                target.removeAttribute('tabindex');
            } else {
                target.setAttribute('tabindex', previousTabindex);
            }
            target.removeAttribute(PERMISSION_TABINDEX_ATTR);
        }
    }
};

const disableElement = (el: HTMLElement) => {
    if (!el.hasAttribute(PERMISSION_POINTER_EVENTS_ATTR)) {
        el.setAttribute(PERMISSION_POINTER_EVENTS_ATTR, el.style.pointerEvents || '');
    }
    el.style.pointerEvents = 'none';
    el.setAttribute('aria-disabled', 'true');
    el.setAttribute(PERMISSION_DISABLED_ATTR, 'true');
    el.classList.add('is-disabled');
    disableNativeControls(el);
};

const enableElement = (el: HTMLElement) => {
    if (!el.hasAttribute(PERMISSION_DISABLED_ATTR)) {
        return;
    }
    el.style.pointerEvents = el.getAttribute(PERMISSION_POINTER_EVENTS_ATTR) || '';
    el.removeAttribute('aria-disabled');
    el.removeAttribute(PERMISSION_DISABLED_ATTR);
    el.removeAttribute(PERMISSION_POINTER_EVENTS_ATTR);
    el.classList.remove('is-disabled');
    enableNativeControls(el);
};

const getComponentPermissionController = (vnode: VNode): PermissionControlledComponent | undefined => {
    return vnode.component?.exposed as PermissionControlledComponent | undefined;
};

const getPermissionMode = (binding: DirectiveBinding<PermissionBindingValue>): PermissionMode => {
    return binding.arg === 'view' ? 'view' : 'manage';
};

const hasNodeAdminDirective = (el: HTMLElement, vnode: VNode) => {
    return el.hasAttribute(NODE_ADMIN_PERMISSION_ATTR) || !!vnode.dirs?.some((item) => item.dir === nodeAdminDirective);
};

const applyPermission = (el: HTMLElement, binding: DirectiveBinding<PermissionBindingValue>, vnode: VNode) => {
    const options = { nodeAdmin: hasNodeAdminDirective(el, vnode) };
    const disabled =
        getPermissionMode(binding) === 'view'
            ? !hasPermissionAccess(binding.value, options)
            : !hasManagePermissionAccess(binding.value, options);
    const controller = getComponentPermissionController(vnode);

    if (controller?.setPermissionDisabled) {
        controller.setPermissionDisabled(disabled);
        return;
    }

    if (disabled) {
        disableElement(el);
        return;
    }
    enableElement(el);
};

const permissionDirective: Directive<HTMLElement, PermissionBindingValue> = {
    mounted(el, binding, vnode) {
        applyPermission(el, binding, vnode);
    },
    updated(el, binding, vnode) {
        applyPermission(el, binding, vnode);
    },
};

export const nodeAdminDirective: Directive<HTMLElement, boolean | undefined> = {
    mounted(el) {
        el.setAttribute(NODE_ADMIN_PERMISSION_ATTR, 'true');
    },
    updated(el) {
        el.setAttribute(NODE_ADMIN_PERMISSION_ATTR, 'true');
    },
    unmounted(el) {
        el.removeAttribute(NODE_ADMIN_PERMISSION_ATTR);
    },
};

export default permissionDirective;
