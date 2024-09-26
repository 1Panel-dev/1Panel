<template>
    <div class="p-w-400">
        <el-descriptions border :column="1">
            <div v-for="(resource, index) of data" :key="index">
                <el-descriptions-item :label="$t('website.' + resource.type)">{{ resource.name }}</el-descriptions-item>
            </div>
        </el-descriptions>
        <el-form
            ref="changeForm"
            :model="req"
            label-position="left"
            label-width="90px"
            class="mt-5"
            v-if="websiteType === 'static' || websiteType === 'runtime'"
        >
            <el-form-item :label="$t('website.changeDatabase')" prop="databaseID">
                <el-select v-model="req.databaseID" class="w-full" @change="changeDatabase">
                    <el-option v-for="(item, index) in databases" :key="index" :label="item.name" :value="item.id">
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
import { ChangeDatabase, GetWebsiteDatabase, GetWebsiteResource } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
    websiteType: {
        type: String,
        default: '',
    },
});
const data = ref([]);
const req = reactive({
    websiteID: props.id,
    databaseID: 0,
    databaseType: '',
});
const databases = ref([]);

const search = async () => {
    try {
        const res = await GetWebsiteResource(props.id);
        data.value = res.data;
    } catch (error) {}
};

const listDatabases = async () => {
    try {
        const res = await GetWebsiteDatabase();
        databases.value = res.data;
        if (databases.value.length > 0) {
            req.databaseID = databases.value[0].id;
        }
    } catch (error) {}
};

const changeDatabase = () => {
    req.databaseType = databases.value.find((item) => item.id === req.databaseID)?.type;
};

const submit = async () => {
    try {
        await ChangeDatabase(req);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
    } catch (error) {}
};

onMounted(() => {
    console.log('websiteType', props.websiteType);
    search();
    listDatabases();
});
</script>
