import { CellAction, Container } from "@maxhub/max-ui";
import { LuLink, LuBadgeInfo, LuDoorOpen, LuWashingMachine, LuCalendarCheck2 } from 'react-icons/lu'

const titles = [
    { id: 1, Icon: LuWashingMachine, label: 'Стирка', link: '/laundry_resident' },
    { id: 2, Icon: LuDoorOpen, label: 'Комнаты', link: '/room_resident' },
    { id: 3, Icon: LuCalendarCheck2, label: 'Дежурства', link: '/cleaning_resident' },
    { id: 4, Icon: LuBadgeInfo, label: 'Информация', link: '/info_resident' },
    { id: 5, Icon: LuLink, label: 'Ссылки на чаты', link: '/chat-links_resident' },
];

const ButtonsPanel = () => (
    <Container>
        {titles.map((item) => (
            <CellAction
                key={item.id}
                before={<item.Icon style={{ fontSize: '2em' }} />}
                showChevron
                asChild
            >
                <a href={item.link}>
                    {item.label}
                </a>
            </CellAction>
        ))}
    </Container>
);

export default ButtonsPanel;
