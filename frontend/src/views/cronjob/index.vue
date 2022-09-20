<template>
    <div>
        <el-button type="primary" @click="onCreate">
            {{ $t('commons.button.create') }}
        </el-button>
        <el-button type="danger" @click="onBatchDelete">
            {{ $t('commons.button.delete') }}
        </el-button>
        <ComplexTable :pagination-config="paginationConfig" style="margin-top: 20px" :data="data">
            <el-table-column :label="$t('cronjob.taskName')" prop="name" />
            <el-table-column :label="$t('commons.table.status')" prop="status" />
            <el-table-column :label="$t('cronjob.cronSpec')">
                <template #default="{ row }">
                    {{ (row.specType, row.hour + '点' + row.minute + '分 执行') }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('cronjob.retainCopies')" prop="retainCopies" />
            <el-table-column :label="$t('cronjob.targetDir')" prop="targetDir" />
        </ComplexTable>

        <el-dialog @close="search" v-model="backupVisiable" width="50%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('cronjob.createCronTask') }}</span>
                </div>
            </template>
            <el-form :model="form" ref="formRef" label-position="left" :rules="rules" label-width="120px">
                <el-form-item :label="$t('cronjob.taskType')" prop="type">
                    <el-select style="width: 60%" v-model="form.type">
                        <el-option
                            v-for="item in typeOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                </el-form-item>

                <el-form-item :label="$t('cronjob.taskName')" prop="name">
                    <el-input style="width: 60%" clearable v-model="form.name" />
                </el-form-item>

                <el-form-item :label="$t('cronjob.cronSpec')" prop="spec">
                    <el-select style="width: 15%" v-model="form.specType">
                        <el-option
                            v-for="item in specOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                    <el-select
                        v-if="form.specType === 'perWeek'"
                        style="width: 12%; margin-left: 20px"
                        v-model="form.week"
                    >
                        <el-option
                            v-for="item in weekOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                    <el-input
                        v-if="form.specType === 'perMonth' || form.specType === 'perNDay'"
                        style="width: 20%; margin-left: 20px"
                        v-model.number="form.day"
                    >
                        <template #append>{{ $t('cronjob.day') }}</template>
                    </el-input>
                    <el-input
                        v-if="form.specType !== 'perHour' && form.specType !== 'perNMinute'"
                        style="width: 20%; margin-left: 20px"
                        v-model.number="form.hour"
                    >
                        <template #append>{{ $t('cronjob.hour') }}</template>
                    </el-input>
                    <el-input style="width: 20%; margin-left: 20px" v-model.number="form.minute">
                        <template #append>{{ $t('cronjob.minute') }}</template>
                    </el-input>
                </el-form-item>

                <el-form-item :label="$t('cronjob.shellContent')" prop="content">
                    <el-input style="width: 60%" clearable type="textarea" v-model="form.content" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="backupVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="onSubmit(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { ref, reactive } from 'vue';
import { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';

const backupVisiable = ref<boolean>(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const data = ref([
    {
        name: '定时访问百度网站',
        status: '正常',
        specType: '每天',
        hour: 1,
        minute: 30,
        retainCopies: 3,
        targetDir: '阿里云 OSS',
    },
]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 5,
    total: 0,
});

const logSearch = reactive({
    page: 1,
    pageSize: 5,
});

const varifySpec = (rule: any, value: any, callback: any) => {
    switch (form.specType) {
        case 'perMonth':
        case 'perNDay':
            if (!(Number.isInteger(form.day) && Number.isInteger(form.hour) && Number.isInteger(form.minute))) {
                console.log('wqweqwe');
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
        case 'perWeek':
            console.log(form.week, Number.isInteger(form.week));
            if (!(Number.isInteger(form.week) && Number.isInteger(form.hour) && Number.isInteger(form.minute))) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
        case 'perNHour':
            if (!(Number.isInteger(form.hour) && Number.isInteger(form.minute))) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
        case 'perHour':
        case 'perNMinute':
            if (!Number.isInteger(form.minute)) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
    }
};

const rules = reactive({
    name: [Rules.requiredInput],
    type: [Rules.requiredSelect],
    specType: [Rules.requiredSelect],
    spec: [
        { validator: varifySpec, trigger: 'blur', required: true },
        { validator: varifySpec, trigger: 'change', required: true },
    ],
    week: [Rules.requiredSelect],
    day: [Rules.number, { max: 31, min: 1 }],
    hour: [Rules.number, { max: 24, min: 0 }],
    minute: [Rules.number, { max: 60, min: 0 }],
    content: [Rules.requiredInput],
});

const form = reactive({
    name: '',
    type: '',
    specType: 'perMonth',
    spec: '',
    week: 1,
    day: 1,
    hour: 2,
    minute: 3,
    content: '',
});

const typeOptions = ref([
    { label: i18n.global.t('cronjob.shell'), value: 'sheel' },
    { label: i18n.global.t('cronjob.website'), value: 'website' },
    { label: i18n.global.t('cronjob.database'), value: 'database' },
    { label: i18n.global.t('cronjob.syncDate'), value: 'sync' },
    { label: i18n.global.t('cronjob.releaseMemory'), value: 'release' },
    { label: i18n.global.t('cronjob.curl'), value: 'curl' },
]);
const specOptions = ref([
    { label: i18n.global.t('cronjob.perMonth'), value: 'perMonth' },
    { label: i18n.global.t('cronjob.perWeek'), value: 'perWeek' },
    { label: i18n.global.t('cronjob.perNDay'), value: 'perNDay' },
    { label: i18n.global.t('cronjob.perNHour'), value: 'perNHour' },
    { label: i18n.global.t('cronjob.perHour'), value: 'perHour' },
    { label: i18n.global.t('cronjob.perNMinute'), value: 'perNMinute' },
]);
const weekOptions = ref([
    { label: i18n.global.t('cronjob.monday'), value: 1 },
    { label: i18n.global.t('cronjob.tuesday'), value: 2 },
    { label: i18n.global.t('cronjob.wednesday'), value: 3 },
    { label: i18n.global.t('cronjob.thursday'), value: 4 },
    { label: i18n.global.t('cronjob.friday'), value: 5 },
    { label: i18n.global.t('cronjob.saturday'), value: 6 },
    { label: i18n.global.t('cronjob.sunday'), value: 7 },
]);
const search = async () => {
    logSearch.page = paginationConfig.currentPage;
    logSearch.pageSize = paginationConfig.pageSize;
    console.log(logSearch);
};

const onCreate = async () => {
    backupVisiable.value = true;
};
const onBatchDelete = async () => {};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        console.log(form);
    });
};
</script>
