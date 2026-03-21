<template>
    <div v-loading="loading">
        <div v-if="!mailStatus.enabled" class="mt-5">
            <LayoutContent :title="$t('menu.mail')">
                <template #main>
                    <div class="flex items-center justify-center" style="height: 300px">
                        <div class="text-center">
                            <el-icon :size="60" color="#409eff"><Message /></el-icon>
                            <p class="mt-4 mb-2">{{ $t('mail.setupHelper') }}</p>
                            <el-button type="primary" @click="onSetup" :loading="setupLoading">
                                {{ $t('mail.setup') }}
                            </el-button>
                        </div>
                    </div>
                </template>
            </LayoutContent>
        </div>

        <div v-else>
            <LayoutContent :title="$t('menu.mail')">
                <template #leftToolBar>
                    <el-button type="primary" @click="onCreate">
                        {{ $t('mail.createDomain') }}
                    </el-button>
                </template>
                <template #rightToolBar>
                    <el-tag v-if="mailStatus.mailContainerUp" type="success" class="mr-2">
                        {{ $t('mail.mailRunning') }}
                    </el-tag>
                    <el-tag v-else type="danger" class="mr-2">
                        {{ $t('mail.mailNotRunning') }}
                    </el-tag>
                    <el-tag v-if="mailStatus.roundcubeUp" type="success" class="mr-2">
                        {{ $t('mail.roundcubeRunning') }}
                    </el-tag>
                    <TableRefresh @search="loadDomains()" />
                </template>
                <template #main>
                    <ComplexTable :data="domains">
                        <el-table-column :label="$t('mail.domainName')" prop="name" min-width="180">
                            <template #default="{ row }">
                                <el-link type="primary" @click="onDetail(row)">{{ row.name }}</el-link>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('mail.accountCount')" prop="accountCount" min-width="100" />
                        <el-table-column :label="$t('mail.aliasCount')" prop="aliasCount" min-width="100" />
                        <el-table-column :label="$t('mail.status')" prop="status" min-width="80">
                            <template #default="{ row }">
                                <el-tag v-if="row.status === 'active'" type="success" size="small">{{ $t('mail.enabled') }}</el-tag>
                                <el-tag v-else type="info" size="small">{{ $t('mail.disabled') }}</el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('commons.table.createdAt')" prop="createdAt" min-width="150">
                            <template #default="{ row }">
                                {{ dateFormat(0, 0, row.createdAt) }}
                            </template>
                        </el-table-column>
                        <fu-table-operations :ellipsis="10" :buttons="buttons" :label="$t('commons.table.operate')" fixed="right" fix />
                    </ComplexTable>
                </template>
            </LayoutContent>
        </div>

        <CreateDomain ref="createRef" @search="loadDomains()" />
        <OpDialog ref="opRef" @search="loadDomains()" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { listMailDomains, getMailStatus, setupMail, deleteMailDomain } from '@/api/modules/mail';
import { Mail } from '@/api/interface/mail';
import { dateFormat } from '@/utils/util';
import CreateDomain from './create.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const router = useRouter();
const loading = ref(false);
const setupLoading = ref(false);
const domains = ref<Mail.DomainRes[]>([]);
const createRef = ref();
const opRef = ref();

const mailStatus = ref<Mail.Status>({
    enabled: false, mailContainerUp: false, mailContainerName: '',
    roundcubeUp: false, roundcubeURL: '', version: '',
});

const buttons = [
    {
        label: i18n.global.t('mail.viewDomain'),
        click: (row: Mail.DomainRes) => onDetail(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Mail.DomainRes) => onDelete(row),
    },
];

const loadDomains = async () => {
    loading.value = true;
    try {
        const res = await listMailDomains();
        domains.value = res.data || [];
    } finally {
        loading.value = false;
    }
};

const loadStatus = async () => {
    try {
        const res = await getMailStatus();
        mailStatus.value = res.data;
    } catch (e) {
        mailStatus.value.enabled = false;
    }
};

const onSetup = async () => {
    setupLoading.value = true;
    try {
        await setupMail({ enable: true });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await loadStatus();
        if (mailStatus.value.enabled) loadDomains();
    } finally {
        setupLoading.value = false;
    }
};

const onCreate = () => createRef.value.acceptParams();
const onDetail = (row: Mail.DomainRes) => router.push({ name: 'MailDomainDetail', params: { id: row.id } });

const onDelete = (row: Mail.DomainRes) => {
    opRef.value.acceptParams({
        title: i18n.global.t('mail.deleteDomain'),
        names: [row.name],
        msg: i18n.global.t('mail.deleteDomainHelper'),
        api: deleteMailDomain,
        params: { id: row.id, cleanupDns: false },
    });
};

onMounted(async () => {
    await loadStatus();
    if (mailStatus.value.enabled) loadDomains();
});
</script>
