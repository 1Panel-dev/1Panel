<template>
    <div class="name-row">
        <div class="name-main" :class="{ 'name-main--actions-only': hideName && !isEditing }">
            <el-form :model="formData" :rules="rules" ref="formRef" v-if="isEditing" @submit.prevent>
                <el-form-item prop="domainName" class="inline-form-item">
                    <el-input
                        v-model="formData.domainName"
                        @keyup.enter.prevent="saveEdit"
                        @blur="saveEdit"
                        @keyup.esc="cancelEdit"
                        class="domain-input"
                        ref="inputRef"
                    />
                </el-form-item>
            </el-form>
            <el-tooltip
                v-else-if="!hideName"
                effect="dark"
                placement="bottom-start"
                popper-class="website-domain-tooltip"
                :show-after="300"
                :disabled="!shouldShowDomainTooltip(row.primaryDomain)"
            >
                <template #content>
                    <div class="website-domain-tooltip__content">
                        {{ getDisplayDomain(row.primaryDomain) }}
                    </div>
                </template>
                <el-text type="primary" class="cursor-pointer domain-text" @click="openConfig(row.id)">
                    <span class="domain-text__content">
                        {{ row.primaryDomain }}
                        <span class="text-gray-400" v-if="isPunycoded(row.primaryDomain)">
                            ({{ GetPunyCodeDomain(row.primaryDomain) }})
                        </span>
                    </span>
                </el-text>
            </el-tooltip>
            <el-popover
                placement="right"
                trigger="hover"
                :width="popoverWidth"
                @before-enter="searchDomains(row.id)"
                v-if="row.type != 'stream'"
            >
                <template #reference>
                    <el-button link icon="Promotion" class="ml-2.5"></el-button>
                </template>
                <table>
                    <tbody>
                        <tr v-for="(domain, index) in domains" :key="index">
                            <td>
                                <el-button type="primary" link @click="openUrl(getUrl(domain, row))">
                                    {{ getUrl(domain, row) }}
                                </el-button>
                            </td>
                            <td>
                                <CopyButton :content="getUrl(domain, row)" />
                            </td>
                        </tr>
                    </tbody>
                </table>
            </el-popover>
            <el-button v-permission link icon="edit" class="ml-2.5" @click="startEdit" v-if="!isEditing"></el-button>
        </div>
        <div v-if="showFavorite">
            <el-tooltip
                effect="dark"
                :content="row.favorite ? $t('commons.table.unpin') : $t('commons.table.pin')"
                placement="top-start"
            >
                <el-button
                    v-permission
                    class="website-pin-button"
                    :class="{ 'is-pinned': row.favorite }"
                    link
                    :size="hideName ? 'default' : 'large'"
                    :type="row.favorite ? 'warning' : 'info'"
                    @click="favoriteWebsite(row)"
                >
                    <svg-icon iconName="p-pushpin" className="website-pin-icon" />
                </el-button>
            </el-tooltip>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, nextTick, computed } from 'vue';
import { listDomains } from '@/api/modules/website';
import { Website } from '@/api/interface/website';
import { routerToNameWithParams } from '@/utils/router';
import { Rules } from '@/global/form-rules';
import { GetPunyCodeDomain, isPunycoded } from '@/utils/misc';
interface Props {
    row: Website.Website;
    defaultHttpPort: number;
    defaultHttpsPort: number;
    hideName?: boolean;
    showFavorite?: boolean;
}
const props = withDefaults(defineProps<Props>(), {
    hideName: false,
    showFavorite: true,
});
const emit = defineEmits(['favoriteChange', 'domainEdit']);
const inputRef = ref();
const isEditing = ref(false);
const domains = ref<Website.Domain[]>([]);
const formData = reactive({
    domainName: '',
});
const rules = ref({
    domainName: [Rules.requiredInput, Rules.linuxName],
});
const formRef = ref();

const ipv6Regex =
    /^(?:(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,7}:|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:(?:(?::[0-9a-fA-F]{1,4}){1,6})|:(?:(?::[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(?::[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(?:ffff(?::0{1,4}){0,1}:){0,1}(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])|(?:[0-9a-fA-F]{1,4}:){1,4}:(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$/;

const isIPv6 = (domain: string): boolean => {
    const cleanDomain = domain.replace(/^\[|\]$/g, '');
    return ipv6Regex.test(cleanDomain);
};

const popoverWidth = computed(() => {
    if (domains.value.length === 0) return 300;

    let maxLength = 0;
    domains.value.forEach((domain) => {
        const url = getUrl(domain, props.row);
        maxLength = Math.max(maxLength, url.length);
    });

    const calculatedWidth = 200 + maxLength * 8 + 60 + 40;

    return Math.min(Math.max(calculatedWidth, 300), 800);
});

const startEdit = () => {
    formData.domainName = props.row.primaryDomain;
    isEditing.value = true;
    nextTick(() => {
        inputRef.value?.focus();
        inputRef.value?.select();
    });
};

const saveEdit = async () => {
    await formRef.value.validate((valid) => {
        if (valid) {
            const editValue = formData.domainName.trim();
            if (editValue && editValue !== props.row.primaryDomain) {
                emit('domainEdit', props.row, editValue);
            }
            isEditing.value = false;
        }
    });
};

const cancelEdit = () => {
    formData.domainName = props.row.primaryDomain;
    isEditing.value = false;
};

const openConfig = (id: number) => {
    routerToNameWithParams('WebsiteConfig', { id: id, tab: 'basic' });
};

const searchDomains = (id: number) => {
    listDomains(id).then((res) => {
        domains.value = res.data;
    });
};

const openUrl = (url: string) => {
    window.open(url);
};

const getUrl = (domain: Website.Domain, website: Website.Website): string => {
    const protocol = website.protocol.toLowerCase();
    let domainStr = domain.domain;

    const cleanDomain = domainStr.replace(/^\[|\]$/g, '');

    if (isIPv6(cleanDomain)) {
        domainStr = `[${cleanDomain}]`;
    }

    let url = `${protocol}://${domainStr}`;

    if (protocol === 'http' && domain.port && domain.port !== 80) {
        url = `${url}:${domain.port}`;
    } else if (protocol === 'https') {
        let port = domain.port;
        if (!domain.ssl) {
            port = props.defaultHttpsPort || 443;
        }
        if (port && port !== 443) {
            url = `${url}:${port}`;
        }
    }

    return url;
};

const favoriteWebsite = (row: Website.Website) => {
    emit('favoriteChange', row);
};

const getDisplayDomain = (domain: string) => {
    if (!isPunycoded(domain)) {
        return domain;
    }
    return `${domain} (${GetPunyCodeDomain(domain)})`;
};

const shouldShowDomainTooltip = (domain: string) => {
    return getDisplayDomain(domain).length > 30;
};
</script>

<style lang="css" scoped>
.name-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    width: 100%;
}

.name-main {
    display: flex;
    align-items: center;
    flex: 1;
    min-width: 0;
}

.name-main--actions-only {
    flex: initial;
}

.domain-text {
    display: inline-flex;
    align-items: center;
    flex: 1;
    width: 0;
    min-width: 0;
    max-width: 100%;
}

.domain-text__content {
    display: inline-block;
    min-width: 0;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

:deep(.el-form) {
    margin: 0;
    line-height: 1;
}
:deep(.el-form-item) {
    margin-bottom: 0;
}
:deep(.el-form-item__error) {
    position: absolute;
    top: 100%;
    left: 0;
    padding-top: 2px;
}

.domain-input {
    width: 200px;
}

:deep(.website-domain-tooltip) {
    max-width: min(720px, calc(100vw - 120px));
}

.website-domain-tooltip__content {
    white-space: normal;
    word-break: break-all;
    line-height: 1.5;
}

.website-pin-button {
    opacity: 0.72;
    transition: opacity 0.2s;
}

.website-pin-button:hover,
.website-pin-button:focus-visible,
.website-pin-button.is-pinned {
    opacity: 1;
}

.website-pin-button :deep(.website-pin-icon) {
    width: 1em;
    height: 1em;
    padding: 0;
    vertical-align: middle;
}
</style>
