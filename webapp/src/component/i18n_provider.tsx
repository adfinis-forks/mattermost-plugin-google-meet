import type {FC} from 'react';
import React from 'react';
import {IntlProvider} from 'react-intl';

import {getTranslations} from '@/plugin/translation';

interface Props {
    currentLocale: string;
    children: React.ReactNode;
}

export const I18nProvider: FC<Props> = ({children, currentLocale}) => {
    return (
        <IntlProvider
            locale={currentLocale}
            key={currentLocale}
            messages={getTranslations(currentLocale)}
        >
            {children}
        </IntlProvider>
    );
};
