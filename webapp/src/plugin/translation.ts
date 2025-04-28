// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import de from 'i18n/de_DE.json';
import en from 'i18n/en.json';
import fr from 'i18n/fr_FR.json';
import pt from 'i18n/pt_BR.json';

export const getTranslations = (locale: string) => {
    switch (locale) {
    case 'pt':
        return pt;
    case 'pt-BR':
        return pt;
    case 'fr':
        return fr;
    case 'fr-FR':
        return fr;
    case 'fr-BE':
        return fr;
    case 'fr-CA':
        return fr;
    case 'de':
        return de;
    case 'de-AT':
        return de;
    case 'de-CH':
        return de;
    case 'de-DE':
        return de;
    case 'de-LI':
        return de;
    case 'de-LU':
        return de;
    }
    return en;
};
