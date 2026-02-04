<template>
    <div class="name-row">
        <div>
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
            <el-text v-else type="primary" class="cursor-pointer" @click="openConfig(row.id)">
                {{ row.primaryDomain }}
                <span class="text-gray-400" v-if="isPunycoded(row.primaryDomain)">
                    ({{ GetPunyCodeDomain(row.primaryDomain) }})
                </span>
            </el-text>
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
            <el-button link icon="edit" class="ml-2.5" @click="startEdit" v-if="!isEditing"></el-button>
        </div>
        <div>
            <el-tooltip effect="dark" :content="$t('website.cancelFavorite')" placement="top-start" v-if="row.favorite">
                <el-button link size="large" icon="StarFilled" type="warning" @click="favoriteWebsite(row)"></el-button>
            </el-tooltip>

            <el-tooltip
                effect="dark"
                :content="$t('website.favorite')"
                placement="top-start"
                v-if="!row.favorite && isHovered"
            >
                <el-button link icon="Star" type="info" @click="favoriteWebsite(row)"></el-button>
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
import { GetPunyCodeDomain, isPunycoded } from '@/utils/util';

interface Props {
    row: Website.Website;
    isHovered: boolean;
    defaultHttpPort: number;
    defaultHttpsPort: number;
}
const props = defineProps<Props>();
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
</script>

<style lang="css" scoped>
.name-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
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
</style>
