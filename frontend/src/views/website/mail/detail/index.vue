<template>
    <div v-loading="loading">
        <LayoutContent :title="domain.name" v-if="domain.id">
            <template #leftToolBar>
                <el-button
                    v-if="mailStatus.roundcubeUp && mailStatus.roundcubeURL"
                    @click="openWebmail"
                    type="primary"
                    plain
                >
                    {{ $t('mail.openWebmail') }}
                </el-button>
            </template>
            <template #main>
                <el-tabs v-model="activeTab" type="border-card">
                    <el-tab-pane :label="$t('mail.accounts')" name="accounts">
                        <Accounts :domain-id="domain.id" :domain-name="domain.name" />
                    </el-tab-pane>
                    <el-tab-pane :label="$t('mail.aliases')" name="aliases">
                        <Aliases :domain-id="domain.id" :domain-name="domain.name" />
                    </el-tab-pane>
                    <el-tab-pane :label="$t('mail.dnsConfig')" name="dns">
                        <DnsConfig :domain="domain" />
                    </el-tab-pane>
                </el-tabs>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { Mail } from '@/api/interface/mail';
import { getMailStatus } from '@/api/modules/mail';
import { listMailDomains } from '@/api/modules/mail';
import Accounts from './accounts.vue';
import Aliases from './aliases.vue';
import DnsConfig from './dns-config.vue';

const route = useRoute();
const loading = ref(false);
const activeTab = ref('accounts');

const domain = ref<Mail.DomainRes>({
    id: 0, name: '', status: '', dkimKey: '', dnsAutoGen: false,
    accountCount: 0, aliasCount: 0,
});

const mailStatus = ref<Mail.Status>({
    enabled: false, mailContainerUp: false, mailContainerName: '',
    roundcubeUp: false, roundcubeURL: '', version: '',
});

const loadDomain = async () => {
    const id = Number(route.params.id);
    if (!id) return;
    const res = await listMailDomains();
    const found = (res.data || []).find((d: Mail.DomainRes) => d.id === id);
    if (found) domain.value = found;
};

const openWebmail = () => {
    if (mailStatus.value.roundcubeURL) {
        window.open(mailStatus.value.roundcubeURL, '_blank');
    }
};

onMounted(async () => {
    loading.value = true;
    try {
        await loadDomain();
        const statusRes = await getMailStatus();
        mailStatus.value = statusRes.data;
    } finally {
        loading.value = false;
    }
});
</script>
