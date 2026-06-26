type FuLocaleMessage = Record<string, unknown>;

const fuLocales: Record<string, FuLocaleMessage> = {
    en: {
        fu: {
            table: {
                more: 'More',
                custom_table_rows: 'Custom columns',
            },
            steps: {
                cancel: 'Cancel',
                prev: 'Previous',
                next: 'Next',
                finish: 'Finish',
            },
        },
    },
    'es-ES': {
        fu: {
            table: {
                more: 'Más',
                custom_table_rows: 'Columnas personalizadas',
            },
            steps: {
                cancel: 'Cancelar',
                prev: 'Anterior',
                next: 'Siguiente',
                finish: 'Finalizar',
            },
        },
    },
    ja: {
        fu: {
            table: {
                more: 'もっと',
                custom_table_rows: 'カスタム列',
            },
            steps: {
                cancel: 'キャンセル',
                prev: '前へ',
                next: '次へ',
                finish: '完了',
            },
        },
    },
    ko: {
        fu: {
            table: {
                more: '더보기',
                custom_table_rows: '사용자 지정 열',
            },
            steps: {
                cancel: '취소',
                prev: '이전',
                next: '다음',
                finish: '완료',
            },
        },
    },
    ms: {
        fu: {
            table: {
                more: 'Lagi',
                custom_table_rows: 'Lajur tersuai',
            },
            steps: {
                cancel: 'Batal',
                prev: 'Sebelumnya',
                next: 'Seterusnya',
                finish: 'Selesai',
            },
        },
    },
    'pt-BR': {
        fu: {
            table: {
                more: 'Mais',
                custom_table_rows: 'Colunas personalizadas',
            },
            steps: {
                cancel: 'Cancelar',
                prev: 'Anterior',
                next: 'Próximo',
                finish: 'Concluir',
            },
        },
    },
    ru: {
        fu: {
            table: {
                more: 'Еще',
                custom_table_rows: 'Настройка колонок',
            },
            steps: {
                cancel: 'Отмена',
                prev: 'Назад',
                next: 'Далее',
                finish: 'Готово',
            },
        },
    },
    tr: {
        fu: {
            table: {
                more: 'Daha fazla',
                custom_table_rows: 'Özel sütunlar',
            },
            steps: {
                cancel: 'İptal',
                prev: 'Önceki',
                next: 'Sonraki',
                finish: 'Tamamla',
            },
        },
    },
    zh: {
        fu: {
            table: {
                more: '更多',
                custom_table_rows: '自定义列',
            },
            steps: {
                cancel: '取消',
                prev: '上一步',
                next: '下一步',
                finish: '完成',
            },
        },
    },
    'zh-Hant': {
        fu: {
            table: {
                more: '更多',
                custom_table_rows: '自訂欄位',
            },
            steps: {
                cancel: '取消',
                prev: '上一步',
                next: '下一步',
                finish: '完成',
            },
        },
    },
    fa: {
        fu: {
            table: {
                more: 'بیشتر',
                custom_table_rows: 'ستون‌های سفارشی',
            },
            steps: {
                cancel: 'لغو',
                prev: 'قبلی',
                next: 'بعدی',
                finish: 'پایان',
            },
        },
    },
};

export const getFuLocaleMessage = (locale: string) => {
    return fuLocales[locale] || fuLocales.en;
};
