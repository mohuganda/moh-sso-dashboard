CREATE TABLE sub_counties (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  district_id UUID NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
  code TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(name, district_id)
);

INSERT INTO
  sub_counties (name, district_id, code)
VALUES
  (
    'Abim DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Abim'
    ),
    NULL
  ),
  (
    'Adjumani DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Adjumani'
    ),
    NULL
  ),
  (
    'Agago DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Agago'
    ),
    NULL
  ),
  (
    'Alebtong DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Alebtong'
    ),
    NULL
  ),
  (
    'Amolatar DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Amolatar'
    ),
    NULL
  ),
  (
    'Amudat DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Amudat'
    ),
    NULL
  ),
  (
    'Amuria DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Amuria'
    ),
    NULL
  ),
  (
    'Amuru DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Amuru'
    ),
    NULL
  ),
  (
    'Apac DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Apac'
    ),
    NULL
  ),
  (
    'Apac Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Apac'
    ),
    NULL
  ),
  (
    'Arua DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Arua'
    ),
    NULL
  ),
  (
    'Arua City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Arua City'
    ),
    NULL
  ),
  (
    'Budaka DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Budaka'
    ),
    NULL
  ),
  (
    'Bududa DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bududa'
    ),
    NULL
  ),
  (
    'Bugiri DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bugiri'
    ),
    NULL
  ),
  (
    'Bugiri Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bugiri'
    ),
    NULL
  ),
  (
    'Bugweri DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bugweri'
    ),
    NULL
  ),
  (
    'Buhweju DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Buhweju'
    ),
    NULL
  ),
  (
    'Buikwe DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Buikwe'
    ),
    NULL
  ),
  (
    'Lugazi Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Buikwe'
    ),
    NULL
  ),
  (
    'Njeru Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Buikwe'
    ),
    NULL
  ),
  (
    'Bukedea DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bukedea'
    ),
    NULL
  ),
  (
    'Bukomansimbi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bukomansimbi'
    ),
    NULL
  ),
  (
    'Bukwo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bukwo'
    ),
    NULL
  ),
  (
    'Bulambuli DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bulambuli'
    ),
    NULL
  ),
  (
    'Buliisa DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Buliisa'
    ),
    NULL
  ),
  (
    'Bundibugyo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bundibugyo'
    ),
    NULL
  ),
  (
    'Bunyangabu DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bunyangabu'
    ),
    NULL
  ),
  (
    'Bushenyi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bushenyi'
    ),
    NULL
  ),
  (
    'Bushenyi-Ishaka Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Bushenyi'
    ),
    NULL
  ),
  (
    'Busia DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Busia'
    ),
    NULL
  ),
  (
    'Busia Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Busia'
    ),
    NULL
  ),
  (
    'Butaleja DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Butaleja'
    ),
    NULL
  ),
  (
    'Butambala DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Butambala'
    ),
    NULL
  ),
  (
    'Butebo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Butebo'
    ),
    NULL
  ),
  (
    'Buvuma DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Buvuma'
    ),
    NULL
  ),
  (
    'Buyende DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Buyende'
    ),
    NULL
  ),
  (
    'Dokolo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Dokolo'
    ),
    NULL
  ),
  (
    'Fort Portal City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Fort Portal City'
    ),
    NULL
  ),
  (
    'Gomba DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Gomba'
    ),
    NULL
  ),
  (
    'Gulu DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Gulu'
    ),
    NULL
  ),
  (
    'Gulu City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Gulu City'
    ),
    NULL
  ),
  (
    'Hoima DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Hoima'
    ),
    NULL
  ),
  (
    'Hoima City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Hoima City'
    ),
    NULL
  ),
  (
    'Ibanda DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Ibanda'
    ),
    NULL
  ),
  (
    'Ibanda Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Ibanda'
    ),
    NULL
  ),
  (
    'Iganga DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Iganga'
    ),
    NULL
  ),
  (
    'Iganga Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Iganga'
    ),
    NULL
  ),
  (
    'Isingiro DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Isingiro'
    ),
    NULL
  ),
  (
    'Jinja DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Jinja'
    ),
    NULL
  ),
  (
    'Jinja City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Jinja City'
    ),
    NULL
  ),
  (
    'Kaabong DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kaabong'
    ),
    NULL
  ),
  (
    'Kabale DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kabale'
    ),
    NULL
  ),
  (
    'Kabale Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kabale'
    ),
    NULL
  ),
  (
    'Kabarole DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kabarole'
    ),
    NULL
  ),
  (
    'Kaberamaido DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kaberamaido'
    ),
    NULL
  ),
  (
    'Kagadi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kagadi'
    ),
    NULL
  ),
  (
    'Kakumiro DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kakumiro'
    ),
    NULL
  ),
  (
    'Kalaki DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kalaki'
    ),
    NULL
  ),
  (
    'Kalangala DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kalangala'
    ),
    NULL
  ),
  (
    'Kaliro DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kaliro'
    ),
    NULL
  ),
  (
    'Kalungu DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kalungu'
    ),
    NULL
  ),
  (
    'Kampala City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kampala'
    ),
    NULL
  ),
  (
    'Kamuli DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kamuli'
    ),
    NULL
  ),
  (
    'Kamuli Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kamuli'
    ),
    NULL
  ),
  (
    'Kamwenge DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kamwenge'
    ),
    NULL
  ),
  (
    'Kanungu DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kanungu'
    ),
    NULL
  ),
  (
    'Kapchorwa DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kapchorwa'
    ),
    NULL
  ),
  (
    'Kapchorwa Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kapchorwa'
    ),
    NULL
  ),
  (
    'Kapelebyong DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kapelebyong'
    ),
    NULL
  ),
  (
    'Karenga DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Karenga'
    ),
    NULL
  ),
  (
    'Kasese DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kasese'
    ),
    NULL
  ),
  (
    'Kasese Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kasese'
    ),
    NULL
  ),
  (
    'Kassanda DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kassanda'
    ),
    NULL
  ),
  (
    'Katakwi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Katakwi'
    ),
    NULL
  ),
  (
    'Kayunga DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kayunga'
    ),
    NULL
  ),
  (
    'Kazo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kazo'
    ),
    NULL
  ),
  (
    'Kibaale DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kibaale'
    ),
    NULL
  ),
  (
    'Kiboga DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kiboga'
    ),
    NULL
  ),
  (
    'Kibuku DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kibuku'
    ),
    NULL
  ),
  (
    'Kikuube DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kikuube'
    ),
    NULL
  ),
  (
    'Kiruhura DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kiruhura'
    ),
    NULL
  ),
  (
    'Kiryandongo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kiryandongo'
    ),
    NULL
  ),
  (
    'Kisoro DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kisoro'
    ),
    NULL
  ),
  (
    'Kisoro Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kisoro'
    ),
    NULL
  ),
  (
    'Kitagwenda DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kitagwenda'
    ),
    NULL
  ),
  (
    'Kitgum DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kitgum'
    ),
    NULL
  ),
  (
    'Kitgum Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kitgum'
    ),
    NULL
  ),
  (
    'Koboko DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Koboko'
    ),
    NULL
  ),
  (
    'Koboko Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Koboko'
    ),
    NULL
  ),
  (
    'Kole DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kole'
    ),
    NULL
  ),
  (
    'Kotido DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kotido'
    ),
    NULL
  ),
  (
    'Kotido Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kotido'
    ),
    NULL
  ),
  (
    'Kumi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kumi'
    ),
    NULL
  ),
  (
    'Kumi Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kumi'
    ),
    NULL
  ),
  (
    'Kwania DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kwania'
    ),
    NULL
  ),
  (
    'Kween DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kween'
    ),
    NULL
  ),
  (
    'Kyankwanzi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kyankwanzi'
    ),
    NULL
  ),
  (
    'Kyegegwa DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kyegegwa'
    ),
    NULL
  ),
  (
    'Kyenjojo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kyenjojo'
    ),
    NULL
  ),
  (
    'Kyotera DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Kyotera'
    ),
    NULL
  ),
  (
    'Lamwo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Lamwo'
    ),
    NULL
  ),
  (
    'Lira DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Lira'
    ),
    NULL
  ),
  (
    'Lira City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Lira City'
    ),
    NULL
  ),
  (
    'Luuka DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Luuka'
    ),
    NULL
  ),
  (
    'Luwero DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Luwero'
    ),
    NULL
  ),
  (
    'Lwengo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Lwengo'
    ),
    NULL
  ),
  (
    'Lyantonde DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Lyantonde'
    ),
    NULL
  ),
  (
    'Madi-Okollo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Madi-Okollo'
    ),
    NULL
  ),
  (
    'Manafwa DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Manafwa'
    ),
    NULL
  ),
  (
    'Maracha DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Maracha'
    ),
    NULL
  ),
  (
    'Masaka DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Masaka'
    ),
    NULL
  ),
  (
    'Masaka City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Masaka City'
    ),
    NULL
  ),
  (
    'Masindi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Masindi'
    ),
    NULL
  ),
  (
    'Masindi Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Masindi'
    ),
    NULL
  ),
  (
    'Mayuge DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mayuge'
    ),
    NULL
  ),
  (
    'Mbale DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mbale'
    ),
    NULL
  ),
  (
    'Mbale City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mbale City'
    ),
    NULL
  ),
  (
    'Mbarara DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mbarara'
    ),
    NULL
  ),
  (
    'Mbarara City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mbarara City'
    ),
    NULL
  ),
  (
    'Mitooma DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mitooma'
    ),
    NULL
  ),
  (
    'Mityana DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mityana'
    ),
    NULL
  ),
  (
    'Mityana Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mityana'
    ),
    NULL
  ),
  (
    'Moroto DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Moroto'
    ),
    NULL
  ),
  (
    'Moroto Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Moroto'
    ),
    NULL
  ),
  (
    'Moyo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Moyo'
    ),
    NULL
  ),
  (
    'Mpigi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mpigi'
    ),
    NULL
  ),
  (
    'Mubende DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mubende'
    ),
    NULL
  ),
  (
    'Mubende Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mubende'
    ),
    NULL
  ),
  (
    'Mukono DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mukono'
    ),
    NULL
  ),
  (
    'Mukono Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Mukono'
    ),
    NULL
  ),
  (
    'Nabilatuk DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Nabilatuk'
    ),
    NULL
  ),
  (
    'Nakapiripirit DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Nakapiripirit'
    ),
    NULL
  ),
  (
    'Nakaseke DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Nakaseke'
    ),
    NULL
  ),
  (
    'Nakasongola DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Nakasongola'
    ),
    NULL
  ),
  (
    'Namayingo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Namayingo'
    ),
    NULL
  ),
  (
    'Namisindwa DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Namisindwa'
    ),
    NULL
  ),
  (
    'Namutumba DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Namutumba'
    ),
    NULL
  ),
  (
    'Napak DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Napak'
    ),
    NULL
  ),
  (
    'Nebbi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Nebbi'
    ),
    NULL
  ),
  (
    'Nebbi Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Nebbi'
    ),
    NULL
  ),
  (
    'Ngora DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Ngora'
    ),
    NULL
  ),
  (
    'Ntoroko DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Ntoroko'
    ),
    NULL
  ),
  (
    'Ntungamo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Ntungamo'
    ),
    NULL
  ),
  (
    'Ntungamo Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Ntungamo'
    ),
    NULL
  ),
  (
    'Nwoya DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Nwoya'
    ),
    NULL
  ),
  (
    'Obongi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Obongi'
    ),
    NULL
  ),
  (
    'Omoro DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Omoro'
    ),
    NULL
  ),
  (
    'Otuke DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Otuke'
    ),
    NULL
  ),
  (
    'Oyam DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Oyam'
    ),
    NULL
  ),
  (
    'Pader DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Pader'
    ),
    NULL
  ),
  (
    'Pakwach DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Pakwach'
    ),
    NULL
  ),
  (
    'Pallisa DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Pallisa'
    ),
    NULL
  ),
  (
    'Rakai DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Rakai'
    ),
    NULL
  ),
  (
    'Rubanda DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Rubanda'
    ),
    NULL
  ),
  (
    'Rubirizi DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Rubirizi'
    ),
    NULL
  ),
  (
    'Rukiga DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Rukiga'
    ),
    NULL
  ),
  (
    'Rukungiri DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Rukungiri'
    ),
    NULL
  ),
  (
    'Rukungiri Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Rukungiri'
    ),
    NULL
  ),
  (
    'Rwampara DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Rwampara'
    ),
    NULL
  ),
  (
    'Sembabule DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Sembabule'
    ),
    NULL
  ),
  (
    'Serere DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Serere'
    ),
    NULL
  ),
  (
    'Sheema DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Sheema'
    ),
    NULL
  ),
  (
    'Sheema Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Sheema'
    ),
    NULL
  ),
  (
    'Sironko DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Sironko'
    ),
    NULL
  ),
  (
    'Soroti DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Soroti'
    ),
    NULL
  ),
  (
    'Soroti City Council',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Soroti City'
    ),
    NULL
  ),
  (
    'Terego DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Terego'
    ),
    NULL
  ),
  (
    'Tororo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Tororo'
    ),
    NULL
  ),
  (
    'Tororo Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Tororo'
    ),
    NULL
  ),
  (
    'Entebbe Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Wakiso'
    ),
    NULL
  ),
  (
    'Kira Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Wakiso'
    ),
    NULL
  ),
  (
    'Makindye Ssabagabo Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Wakiso'
    ),
    NULL
  ),
  (
    'Nansana Municipality',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Wakiso'
    ),
    NULL
  ),
  (
    'Wakiso DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Wakiso'
    ),
    NULL
  ),
  (
    'Yumbe DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Yumbe'
    ),
    NULL
  ),
  (
    'Zombo DLG',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Zombo'
    ),
    NULL
  ),
  (
    'nan',
    (
      SELECT
        id
      FROM
        districts
      WHERE
        name = 'Zombo'
    ),
    NULL
  ) ON CONFLICT (name, district_id) DO
UPDATE
SET
  updated_at = now();