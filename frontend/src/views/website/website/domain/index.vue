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
            </el-text>
            <el-popover
                placement="right"
                trigger="hover"
                :width="300"
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
import { ref } from 'vue';
import { listDomains } from '@/api/modules/website';
import { Website } from '@/api/interface/website';
import { routerToNameWithParams } from '@/utils/router';
import { Rules } from '@/global/form-rules';

interface Props {
    row: Website.Website;
    isHovered: boolean;
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
    let url = protocol + '://' + domain.domain;
    if (protocol == 'http' && domain.port != 80) {
        url = url + ':' + domain.port;
    }
    if (protocol == 'https' && domain.ssl) {
        url = url + ':' + domain.port;
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
