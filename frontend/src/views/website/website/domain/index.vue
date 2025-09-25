<template>
    <div class="name-row">
        <div>
            <el-input
                v-if="isEditing"
                v-model="editValue"
                @keyup.enter="saveEdit"
                @blur="saveEdit"
                @keyup.esc="cancelEdit"
                class="domain-input"
                ref="inputRef"
            />
            <el-text v-else type="primary" class="cursor-pointer" @click="openConfig(row.id)">
                {{ row.primaryDomain }}
            </el-text>
            <el-popover placement="right" trigger="hover" :width="300" @before-enter="searchDomains(row.id)">
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
            <el-button link icon="edit" @click="startEdit" v-if="!isEditing"></el-button>
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

interface Props {
    row: Website.Website;
    isHovered: boolean;
}
const props = defineProps<Props>();
const emit = defineEmits(['favoriteChange', 'domainEdit']);
const inputRef = ref();
const isEditing = ref(false);
const editValue = ref('');
const domains = ref<Website.Domain[]>([]);

const startEdit = () => {
    editValue.value = props.row.primaryDomain;
    isEditing.value = true;
    nextTick(() => {
        inputRef.value?.focus();
        inputRef.value?.select();
    });
};

const saveEdit = () => {
    if (editValue.value.trim() && editValue.value !== props.row.primaryDomain) {
        emit('domainEdit', props.row, editValue.value.trim());
    }
    isEditing.value = false;
};

const cancelEdit = () => {
    editValue.value = props.row.primaryDomain;
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
</style>
