import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NavbardashComponent } from './navbardash.component';

describe('NavbardashComponent', () => {
  let component: NavbardashComponent;
  let fixture: ComponentFixture<NavbardashComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ NavbardashComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(NavbardashComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
