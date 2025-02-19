<template>
    <div class="p-w-400">
        <el-descriptions border :column="1">
            <div v-for="(resource, index) of data" :key="index">
                <el-descriptions-item :label="$t('menu.' + resource.type)">{{ resource.name }}</el-descriptions-item>
            </div>
        </el-descriptions>
        <el-form
            ref="changeForm"
            :model="req"
            label-position="left"
            label-width="90px"
            class="mt-5"
            v-if="website.type === 'static' || website.type === 'runtime'"
        >
            <el-form-item :label="$t('website.changeDatabase')" prop="db">
                <el-select v-model="req.db" class="w-full" @change="changeDB">
                    <el-option
                        v-for="(item, index) in databases"
                        :key="index"
                        :label="item.name"
                        :value="item.id + item.type"
                    >
                        <div class="flex justify-between items-center">
                            <span>{{ item.name }}</span>
                            <el-tag>{{ item.type }}</el-tag>
                        </div>
                    </el-option>
                </el-select>
                <el-text type="warning">{{ $t('website.changeDatabaseHelper1') }}</el-text>
                <el-text type="warning">{{ $t('website.changeDatabaseHelper2') }}</el-text>
            </el-form-item>
            <el-form-item>
                <el-button type="primary" @click="submit()">
                    {{ $t('commons.button.save') }}
                </el-button>
            </el-form-item>
        </el-form>
    </div>
</template>
<script setup lang="ts">
import { changeDatabase, getWebsite, getWebsiteDatabase, getWebsiteResource } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const data = ref([]);
const req = reactive({
    websiteID: props.id,
    databaseID: 0,
    databaseType: '',
    db: '',
});
const databases = ref([]);
const website = ref({
    type: '',
    dbID: 0,
    dbType: '',
});

const search = async () => {
    try {
        const res = await getWebsiteResource(props.id);
        data.value = res.data;
    } catch (error) {}
};

const listDatabases = async () => {
    try {
        const res = await getWebsiteDatabase();
        databases.value = res.data;
        if (databases.value.length > 0) {
            if (website.value.dbID > 0) {
                for (let i = 0; i < databases.value.length; i++) {
                    if (
                        databases.value[i].id === website.value.dbID &&
                        databases.value[i].type === website.value.dbType
                    ) {
                        req.db = databases.value[i].id + databases.value[i].type;
                        break;
                    }
                }
            } else {
                req.db = databases.value[0];
            }
        }
    } catch (error) {}
};

const changeDB = () => {
    for (let i = 0; i < databases.value.length; i++) {
        if (databases.value[i].id + databases.value[i].type === req.db) {
            req.databaseID = databases.value[i].id;
            req.databaseType = databases.value[i].type;
            break;
        }
    }
};

const submit = async () => {
    try {
        await changeDatabase(req);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
    } catch (error) {}
};

const get = async () => {
    try {
        const res = await getWebsite(props.id);
        website.value = res.data;
        req.db = '';
        search();
        listDatabases();
    } catch (error) {}
};

onMounted(() => {
    get();
});
</script>
